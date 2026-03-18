// Package installer provides dependency installation and management.
// This file implements dependency resolution.
package installer

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// DependencyResolver resolves dependency order and checks for conflicts.
type DependencyResolver struct {
	// registry contains all known dependencies.
	registry map[string]*Dependency

	// installed tracks what's installed.
	installed map[string]string // name -> version

	// mu protects concurrent access.
	mu sync.RWMutex
}

// NewDependencyResolver creates a new resolver with built-in dependencies.
func NewDependencyResolver() *DependencyResolver {
	r := &DependencyResolver{
		registry:  make(map[string]*Dependency),
		installed: make(map[string]string),
	}

	// Register built-in dependencies
	r.registerBuiltins()

	return r
}

// registerBuiltins registers all built-in dependencies.
func (r *DependencyResolver) registerBuiltins() {
	// oh-my-posh - primary prompt engine
	r.Register(&Dependency{
		Name:          "oh-my-posh",
		DisplayName:   "Oh My Posh",
		Description:   "Prompt theme engine for any shell",
		Version:       "latest",
		Type:          ComponentTypeBinary,
		Source:        "https://github.com/jandedobbeleer/oh-my-posh/releases/latest/download/posh-{os}-{arch}",
		Dependencies:  []string{},
		Platforms:     []Platform{PlatformMacOS, PlatformLinux, PlatformWindows, PlatformWSL},
		InstallPath:   "~/.local/bin/oh-my-posh",
		Optional:      false,
		VerifyCommand: "oh-my-posh --version",
	})

	// Nerd Fonts - recommended font
	r.Register(&Dependency{
		Name:         "MesloLGM-NF",
		DisplayName:  "MesloLGM Nerd Font",
		Description:  "Patched font with icons and symbols",
		Version:      "latest",
		Type:         ComponentTypeFont,
		Source:       "https://github.com/ryanoasis/nerd-fonts/releases/latest",
		Dependencies: []string{},
		Platforms:    []Platform{PlatformMacOS, PlatformLinux},
		InstallPath:  "~/Library/Fonts or ~/.local/share/fonts",
		Optional:     true,
	})

	// zoxide - smart directory changer
	r.Register(&Dependency{
		Name:          "zoxide",
		DisplayName:   "Zoxide",
		Description:   "Fast directory jumping (z command)",
		Version:       "latest",
		Type:          ComponentTypeBinary,
		Source:        "https://github.com/ajeetdsouza/zoxide/releases/latest/download/zoxide-{os}-{arch}",
		Dependencies:  []string{},
		Platforms:     []Platform{PlatformMacOS, PlatformLinux},
		InstallPath:   "~/.local/bin/zoxide",
		Optional:      true,
		VerifyCommand: "zoxide --version",
	})

	// fzf - fuzzy finder
	r.Register(&Dependency{
		Name:          "fzf",
		DisplayName:   "FZF",
		Description:   "Command-line fuzzy finder",
		Version:       "latest",
		Type:          ComponentTypeBinary,
		Source:        "https://github.com/junegunn/fzf/releases/latest/download/fzf-{os}-{arch}",
		Dependencies:  []string{},
		Platforms:     []Platform{PlatformMacOS, PlatformLinux},
		InstallPath:   "~/.local/bin/fzf",
		Optional:      true,
		VerifyCommand: "fzf --version",
	})

	// bat - better cat
	r.Register(&Dependency{
		Name:          "bat",
		DisplayName:   "Bat",
		Description:   "Cat clone with syntax highlighting",
		Version:       "latest",
		Type:          ComponentTypePackage,
		Source:        "package:bash",
		Dependencies:  []string{},
		Platforms:     []Platform{PlatformMacOS, PlatformLinux},
		Optional:      true,
		VerifyCommand: "bat --version",
	})

	// eza - better ls
	r.Register(&Dependency{
		Name:          "eza",
		DisplayName:   "Eza",
		Description:   "Modern ls replacement with colors",
		Version:       "latest",
		Type:          ComponentTypePackage,
		Source:        "package:bash",
		Dependencies:  []string{},
		Platforms:     []Platform{PlatformMacOS, PlatformLinux},
		Optional:      true,
		VerifyCommand: "eza --version",
	})
}

// Register adds a dependency to the registry.
func (r *DependencyResolver) Register(dep *Dependency) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registry[dep.Name] = dep
}

// GetDependency retrieves a dependency by name.
func (r *DependencyResolver) GetDependency(name string) *Dependency {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.registry[name]
}

// GetAllDependencies returns all registered dependencies.
func (r *DependencyResolver) GetAllDependencies() []*Dependency {
	r.mu.RLock()
	defer r.mu.RUnlock()

	deps := make([]*Dependency, 0, len(r.registry))
	for _, dep := range r.registry {
		deps = append(deps, dep)
	}
	return deps
}

// MarkInstalled marks a dependency as installed.
func (r *DependencyResolver) MarkInstalled(name, version string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.installed[name] = version
}

// IsInstalled checks if a dependency is installed.
func (r *DependencyResolver) IsInstalled(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.installed[name]
	return ok
}

// GetInstalledVersion returns the installed version of a dependency.
func (r *DependencyResolver) GetInstalledVersion(name string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.installed[name]
}

// Resolve resolves dependency order for the given components.
// Returns dependencies in the order they should be installed.
func (r *DependencyResolver) Resolve(names []string) ([]*Dependency, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build dependency graph
	resolved := make([]*Dependency, 0)
	visited := make(map[string]bool)
	visiting := make(map[string]bool) // For cycle detection

	var resolve func(name string) error
	resolve = func(name string) error {
		// Already resolved
		if visited[name] {
			return nil
		}

		// Cycle detected
		if visiting[name] {
			return fmt.Errorf("dependency cycle detected: %s", name)
		}

		// Get dependency from registry
		dep, ok := r.registry[name]
		if !ok {
			return fmt.Errorf("unknown dependency: %s", name)
		}

		visiting[name] = true

		// Resolve dependencies first
		for _, depName := range dep.Dependencies {
			if err := resolve(depName); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		resolved = append(resolved, dep)

		return nil
	}

	// Resolve each requested dependency
	for _, name := range names {
		if err := resolve(name); err != nil {
			return nil, err
		}
	}

	return resolved, nil
}

// CheckAlreadyInstalled checks which dependencies are already installed.
func (r *DependencyResolver) CheckAlreadyInstalled() []*DependencyStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	statuses := make([]*DependencyStatus, 0, len(r.registry))

	for name, dep := range r.registry {
		status := &DependencyStatus{
			Name:      name,
			Installed: r.installed[name] != "",
			Version:   r.installed[name],
			Required:  dep.Version,
		}

		if status.Installed {
			status.Satisfied = true
			status.Missing = false
			status.Outdated = false // TODO: Version comparison
		} else {
			status.Satisfied = false
			status.Missing = true
		}

		statuses = append(statuses, status)
	}

	// Sort by name
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	return statuses
}

// DetectConflicts finds conflicting dependencies.
func (r *DependencyResolver) DetectConflicts(names []string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conflicts := make([]string, 0)

	// Check for platform conflicts
	for _, name := range names {
		dep, ok := r.registry[name]
		if !ok {
			continue
		}

		// Check platform support
		if len(dep.Platforms) > 0 {
			supported := false
			// Platform checking would need context
			// For now, assume all platforms supported
			_ = supported
		}

		// TODO: Check version conflicts
	}

	return conflicts
}

// GeneratePlan creates an installation plan for the given dependencies.
func (r *DependencyResolver) GeneratePlan(names []string, platform Platform) (*InstallPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Resolve dependencies
	resolved, err := r.Resolve(names)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Filter by platform
	filtered := make([]*Dependency, 0)
	for _, dep := range resolved {
		if len(dep.Platforms) == 0 {
			// No platform restriction, works everywhere
			filtered = append(filtered, dep)
			continue
		}

		// Check if platform is supported
		for _, p := range dep.Platforms {
			if p == platform {
				filtered = append(filtered, dep)
				break
			}
		}
	}

	// Calculate totals
	var totalSize int64
	requiresDownload := false

	for _, dep := range filtered {
		// Estimate size (most binaries/fonts are 5-50MB)
		switch dep.Type {
		case ComponentTypeBinary:
			totalSize += 20 * 1024 * 1024 // 20MB estimate
			requiresDownload = true
		case ComponentTypeFont:
			totalSize += 30 * 1024 * 1024 // 30MB estimate
			requiresDownload = true
		}

		if dep.IsInstalled {
			requiresDownload = false
		}
	}

	// Detect conflicts
	conflicts := r.DetectConflicts(names)

	plan := &InstallPlan{
		Components:        filtered,
		TotalSize:         totalSize,
		EstimatedDuration: estimateDuration(len(filtered)),
		RequiresDownload:  requiresDownload,
		RequiresRestart:   true, // Most installations require shell restart
		Conflicts:         conflicts,
		Warnings:          []string{},
	}

	return plan, nil
}

// estimateDuration estimates installation time based on component count.
func estimateDuration(count int) time.Duration {
	// Rough estimate: 30 seconds per component for download + install
	return time.Duration(count*30) * time.Second
}

// GetDependencyOrder returns the installation order for the given dependencies.
// Dependencies are sorted so that required dependencies are installed first.
func (r *DependencyResolver) GetDependencyOrder(names []string) ([]string, error) {
	resolved, err := r.Resolve(names)
	if err != nil {
		return nil, err
	}

	order := make([]string, len(resolved))
	for i, dep := range resolved {
		order[i] = dep.Name
	}

	return order, nil
}

// CheckVersionConflicts checks for version conflicts between dependencies.
func (r *DependencyResolver) CheckVersionConflicts(names []string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conflicts := make([]string, 0)

	// Track required versions
	requiredVersions := make(map[string][]string) // dep -> versions that require it

	for _, name := range names {
		dep, ok := r.registry[name]
		if !ok {
			continue
		}

		for _, depName := range dep.Dependencies {
			requiredVersions[depName] = append(requiredVersions[depName], name)
		}
	}

	// Check if any dependency is required by multiple parents with different versions
	for depName, parents := range requiredVersions {
		// For now, just check if multiple parents
		if len(parents) > 1 {
			conflicts = append(conflicts, fmt.Sprintf("dependency %s required by multiple packages: %s",
				depName, strings.Join(parents, ", ")))
		}
	}

	return conflicts
}

// ClearInstalled clears the installed state for testing.
func (r *DependencyResolver) ClearInstalled() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.installed = make(map[string]string)
}
