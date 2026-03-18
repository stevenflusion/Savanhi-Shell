// Package staging provides change staging for Savanhi Shell installation.
// It queues pending configuration changes and validates them before commit.
package staging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// InstallContext contains installation context information.
type InstallContext struct {
	// ConfigDir is the Savanhi configuration directory.
	ConfigDir string `json:"config_dir"`

	// HomeDir is the user's home directory.
	HomeDir string `json:"home_dir"`

	// BinDir is the binary installation directory.
	BinDir string `json:"bin_dir"`

	// FontDir is the font installation directory.
	FontDir string `json:"font_dir"`

	// CacheDir is the download cache directory.
	CacheDir string `json:"cache_dir"`

	// OS is the detected OS information.
	OS string `json:"os"`

	// Shell is the detected shell.
	Shell string `json:"shell"`
}

// StagingSystem manages pending configuration changes.
type StagingSystem struct {
	// context is the installation context.
	context *InstallContext

	// stagingDir is the staging directory.
	stagingDir string

	// changes are the staged changes.
	changes []*StagedChange

	// mu protects concurrent access.
	mu sync.RWMutex
}

// StagedChange represents a staged configuration change.
type StagedChange struct {
	// ID is the unique identifier for this change.
	ID string `json:"id"`

	// Component is the component being changed.
	Component string `json:"component"`

	// Action is the action being performed (install, configure, remove).
	Action string `json:"action"`

	// Target is the target file or configuration.
	Target string `json:"target"`

	// Before is the state before the change.
	Before interface{} `json:"before,omitempty"`

	// After is the state after the change.
	After interface{} `json:"after"`

	// Dependencies are IDs of changes that must be applied first.
	Dependencies []string `json:"dependencies,omitempty"`

	// CreatedAt is when the change was staged.
	CreatedAt time.Time `json:"created_at"`

	// Status is the current status of the change.
	Status ChangeStatus `json:"status"`

	// Checksum is the checksum before the change.
	Checksum string `json:"checksum,omitempty"`
}

// ChangeStatus represents the status of a staged change.
type ChangeStatus string

const (
	// StatusPending indicates the change is pending.
	StatusPending ChangeStatus = "pending"
	// StatusValidated indicates the change has been validated.
	StatusValidated ChangeStatus = "validated"
	// StatusApplied indicates the change has been applied.
	StatusApplied ChangeStatus = "applied"
	// StatusFailed indicates the change failed.
	StatusFailed ChangeStatus = "failed"
	// StatusRolledBack indicates the change was rolled back.
	StatusRolledBack ChangeStatus = "rolled_back"
)

// ValidationError represents a validation error.
type ValidationError struct {
	// Change is the change that failed validation.
	Change *StagedChange `json:"change"`

	// Field is the field that failed.
	Field string `json:"field"`

	// Message is the error message.
	Message string `json:"message"`
}

// Error returns the error message.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// NewStagingSystem creates a new staging system.
func NewStagingSystem(ctx *InstallContext) *StagingSystem {
	stagingDir := filepath.Join(ctx.ConfigDir, "staging")
	return &StagingSystem{
		context:    ctx,
		stagingDir: stagingDir,
		changes:    make([]*StagedChange, 0),
	}
}

// Queue adds a change to the staging queue.
func (s *StagingSystem) Queue(change *StagedChange) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate ID if not set
	if change.ID == "" {
		change.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Set timestamps
	if change.CreatedAt.IsZero() {
		change.CreatedAt = time.Now()
	}

	// Set initial status
	if change.Status == "" {
		change.Status = StatusPending
	}

	// Check for duplicate
	for _, existing := range s.changes {
		if existing.ID == change.ID {
			return fmt.Errorf("change with ID %s already exists", change.ID)
		}

		// Check for conflicting changes to same target
		if existing.Target == change.Target && existing.Action == change.Action {
			return fmt.Errorf("conflicting change: %s already queued for %s", change.Action, change.Target)
		}
	}

	s.changes = append(s.changes, change)

	// Persist staging state
	return s.persist()
}

// QueueMultiple adds multiple changes to the staging queue.
func (s *StagingSystem) QueueMultiple(changes []*StagedChange) error {
	for _, change := range changes {
		if err := s.Queue(change); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates all staged changes.
func (s *StagingSystem) Validate() []*ValidationError {
	s.mu.RLock()
	defer s.mu.RUnlock()

	errors := make([]*ValidationError, 0)

	for _, change := range s.changes {
		// Validate target path
		if change.Target == "" {
			errors = append(errors, &ValidationError{
				Change:  change,
				Field:   "target",
				Message: "target is required",
			})
			continue
		}

		// Validate action
		switch change.Action {
		case "install", "configure", "remove":
			// Valid actions
		default:
			errors = append(errors, &ValidationError{
				Change:  change,
				Field:   "action",
				Message: fmt.Sprintf("invalid action: %s", change.Action),
			})
		}

		// Validate dependencies exist
		for _, depID := range change.Dependencies {
			if !s.hasChange(depID) {
				errors = append(errors, &ValidationError{
					Change:  change,
					Field:   "dependencies",
					Message: fmt.Sprintf("dependency %s not found", depID),
				})
			}
		}

		// Validate target path is not protected
		if s.isProtectedPath(change.Target) {
			errors = append(errors, &ValidationError{
				Change:  change,
				Field:   "target",
				Message: "target path is protected",
			})
		}

		// Mark as validated if no errors
		if len(errors) == 0 || errors[len(errors)-1].Change != change {
			change.Status = StatusValidated
		}
	}

	return errors
}

// DetectConflicts detects conflicts between staged changes.
func (s *StagingSystem) DetectConflicts() []*Conflict {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conflicts := make([]*Conflict, 0)

	// Check for conflicting targets
	targetChanges := make(map[string][]*StagedChange)
	for _, change := range s.changes {
		targetChanges[change.Target] = append(targetChanges[change.Target], change)
	}

	for target, changes := range targetChanges {
		if len(changes) > 1 {
			conflicts = append(conflicts, &Conflict{
				Target:              target,
				Changes:             changes,
				Reason:              "multiple changes to same target",
				SuggestedResolution: "merge changes or select one",
			})
		}
	}

	// Check for circular dependencies
	for _, change := range s.changes {
		visited := make(map[string]bool)
		if s.hasCircularDependency(change.ID, visited) {
			conflicts = append(conflicts, &Conflict{
				Target:              change.ID,
				Changes:             []*StagedChange{change},
				Reason:              "circular dependency detected",
				SuggestedResolution: "remove circular dependency",
			})
		}
	}

	return conflicts
}

// Conflict represents a detected conflict.
type Conflict struct {
	// Target is the conflicting target or ID.
	Target string `json:"target"`

	// Changes are the conflicting changes.
	Changes []*StagedChange `json:"changes"`

	// Reason is the reason for the conflict.
	Reason string `json:"reason"`

	// SuggestedResolution is a suggested resolution.
	SuggestedResolution string `json:"suggested_resolution"`
}

// GetPending returns all pending changes in dependency order.
func (s *StagingSystem) GetPending() []*StagedChange {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Sort by dependencies
	sorted := s.sortByDependencies()
	pending := make([]*StagedChange, 0)

	for _, change := range sorted {
		if change.Status == StatusPending || change.Status == StatusValidated {
			pending = append(pending, change)
		}
	}

	return pending
}

// GetStaged returns all staged changes.
func (s *StagingSystem) GetStaged() []*StagedChange {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*StagedChange, len(s.changes))
	copy(result, s.changes)
	return result
}

// Commit applies all staged changes.
func (s *StagingSystem) Commit() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate first
	errors := s.validateLocked()
	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %v", errors)
	}

	// Sort by dependencies
	sorted := s.sortByDependencies()

	// Apply each change
	for _, change := range sorted {
		if change.Status == StatusApplied {
			continue
		}

		// Apply change based on action
		if err := s.applyChange(change); err != nil {
			// Mark as failed
			change.Status = StatusFailed

			// Rollback previous changes
			s.rollbackChanges()
			return fmt.Errorf("failed to apply change %s: %w", change.ID, err)
		}

		change.Status = StatusApplied
	}

	// Clear staging
	s.changes = make([]*StagedChange, 0)

	// Persist
	return s.persist()
}

// Clear removes all staged changes without applying.
func (s *StagingSystem) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.changes = make([]*StagedChange, 0)
	return s.persist()
}

// Remove removes a specific staged change.
func (s *StagingSystem) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, change := range s.changes {
		if change.ID == id {
			// Remove from slice
			s.changes = append(s.changes[:i], s.changes[i+1:]...)
			return s.persist()
		}
	}

	return fmt.Errorf("change %s not found", id)
}

// sortByDependencies sorts changes by their dependencies using topological sort.
func (s *StagingSystem) sortByDependencies() []*StagedChange {
	if len(s.changes) == 0 {
		return nil
	}

	// Build dependency graph
	graph := make(map[string][]string)
	changeMap := make(map[string]*StagedChange)

	for _, change := range s.changes {
		graph[change.ID] = change.Dependencies
		changeMap[change.ID] = change
	}

	// Topological sort
	sorted := make([]*StagedChange, 0)
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(id string) error
	visit = func(id string) error {
		if visited[id] {
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("circular dependency detected: %s", id)
		}

		visiting[id] = true

		for _, dep := range graph[id] {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visiting[id] = false
		visited[id] = true

		if change, ok := changeMap[id]; ok {
			sorted = append(sorted, change)
		}

		return nil
	}

	for _, change := range s.changes {
		if err := visit(change.ID); err != nil {
			// Fall back to original order on error
			return s.changes
		}
	}

	return sorted
}

// hasChange checks if a change with the given ID exists.
func (s *StagingSystem) hasChange(id string) bool {
	for _, change := range s.changes {
		if change.ID == id {
			return true
		}
	}
	return false
}

// isProtectedPath checks if a path is protected.
func (s *StagingSystem) isProtectedPath(path string) bool {
	protectedPaths := []string{
		"/etc/passwd",
		"/etc/shadow",
		"/etc/sudoers",
		"/root",
	}

	for _, protected := range protectedPaths {
		if path == protected {
			return true
		}
	}

	return false
}

// hasCircularDependency checks for circular dependencies.
func (s *StagingSystem) hasCircularDependency(id string, visited map[string]bool) bool {
	change := s.getChangeLocked(id)
	if change == nil {
		return false
	}

	if visited[id] {
		return true
	}

	visited[id] = true
	for _, dep := range change.Dependencies {
		if s.hasCircularDependency(dep, visited) {
			return true
		}
	}
	delete(visited, id)

	return false
}

// getChangeLocked gets a change by ID (must hold lock).
func (s *StagingSystem) getChangeLocked(id string) *StagedChange {
	for _, change := range s.changes {
		if change.ID == id {
			return change
		}
	}
	return nil
}

// validateLocked validates changes (must hold lock).
func (s *StagingSystem) validateLocked() []*ValidationError {
	errors := make([]*ValidationError, 0)

	for _, change := range s.changes {
		if change.Target == "" {
			errors = append(errors, &ValidationError{
				Change:  change,
				Field:   "target",
				Message: "target is required",
			})
		}

		if change.Action == "" {
			errors = append(errors, &ValidationError{
				Change:  change,
				Field:   "action",
				Message: "action is required",
			})
		}
	}

	return errors
}

// applyChange applies a single change.
func (s *StagingSystem) applyChange(change *StagedChange) error {
	// Change application is handled by the installation flow
	// This is a placeholder for the staging system
	change.Status = StatusApplied
	return nil
}

// rollbackChanges rolls back all applied changes.
func (s *StagingSystem) rollbackChanges() {
	for i := len(s.changes) - 1; i >= 0; i-- {
		change := s.changes[i]
		if change.Status == StatusApplied {
			change.Status = StatusRolledBack
		}
	}
}

// persist saves the staging state to disk.
func (s *StagingSystem) persist() error {
	if err := os.MkdirAll(s.stagingDir, 0755); err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}

	data, err := json.MarshalIndent(s.changes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal staging state: %w", err)
	}

	path := filepath.Join(s.stagingDir, "pending.json")
	tempPath := path + ".tmp"

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write staging state: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to save staging state: %w", err)
	}

	return nil
}

// Load loads the staging state from disk.
func (s *StagingSystem) Load() error {
	path := filepath.Join(s.stagingDir, "pending.json")

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil // No existing state
	}
	if err != nil {
		return fmt.Errorf("failed to read staging state: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := json.Unmarshal(data, &s.changes); err != nil {
		return fmt.Errorf("failed to unmarshal staging state: %w", err)
	}

	return nil
}

// GetChangeCount returns the number of staged changes.
func (s *StagingSystem) GetChangeCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.changes)
}

// HasPending checks if there are any pending changes.
func (s *StagingSystem) HasPending() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.changes) > 0
}
