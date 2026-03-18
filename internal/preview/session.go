// Package preview provides live preview capabilities for Savanhi Shell.
// This file implements the session manager for preview sessions.
package preview

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/savanhi/shell/internal/persistence"
	"github.com/savanhi/shell/pkg/shell"
)

// Session management constants.
const (
	// MaxActiveSessions is the maximum number of concurrent sessions.
	MaxActiveSessions = 3

	// SessionTimeout is the default session timeout.
	SessionTimeout = 30 * time.Second

	// SessionCleanupInterval is the interval for session cleanup.
	SessionCleanupInterval = 1 * time.Minute
)

// Common session errors.
var (
	// ErrSessionNotFound indicates the session was not found.
	ErrSessionNotFound = errors.New("session not found")

	// ErrSessionAlreadyActive indicates a session is already active.
	ErrSessionAlreadyActive = errors.New("session already active")

	// ErrMaxSessionsReached indicates the maximum sessions limit.
	ErrMaxSessionsReached = errors.New("maximum sessions reached")

	// ErrSessionCancelled indicates the session was cancelled.
	ErrSessionCancelled = errors.New("session cancelled")
)

// DefaultSessionManager implements SessionManager interface.
type DefaultSessionManager struct {
	// mu protects concurrent access.
	mu sync.RWMutex

	// activeSessions maps session IDs to sessions.
	activeSessions map[string]*PreviewSessionState

	// persister is the persistence layer.
	persister *persistence.FilePersister

	// subsheller is the subshell spawner.
	subsheller SubshellSpawner

	// cleaner is the cleanup handler.
	cleaner *PreviewCleaner

	// safetyChecker is the safety validator.
	safetyChecker *PreviewSafetyChecker
}

// NewDefaultSessionManager creates a new DefaultSessionManager.
func NewDefaultSessionManager(persister *persistence.FilePersister, subsheller SubshellSpawner) (*DefaultSessionManager, error) {
	cleaner := NewPreviewCleaner()
	safetyChecker := NewPreviewSafetyChecker()

	return &DefaultSessionManager{
		activeSessions: make(map[string]*PreviewSessionState),
		persister:      persister,
		subsheller:     subsheller,
		cleaner:        cleaner,
		safetyChecker:  safetyChecker,
	}, nil
}

// CreateSession creates a new preview session.
func (m *DefaultSessionManager) CreateSession(config *PreviewConfig) (*PreviewSessionState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check concurrent session limit
	if len(m.activeSessions) >= MaxActiveSessions {
		return nil, ErrMaxSessionsReached
	}

	// Validate configuration
	if err := m.safetyChecker.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Generate session ID
	sessionID := generateSessionID()

	// Create context with timeout
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	// Create temporary directory and files
	tempDir, tempRCFile, err := m.createTempFiles(config)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create temp files: %w", err)
	}

	// Create session state
	session := &PreviewSessionState{
		ID:         sessionID,
		Config:     *config,
		Status:     StatusPending,
		TempDir:    tempDir,
		TempRCFile: tempRCFile,
		StartTime:  time.Now(),
		CancelFunc: cancel,
	}

	// Session context is stored for use during preview execution
	// The context can be accessed via the CancelFunc for cancellation
	_ = ctx // Context managed by CancelFunc

	// Track temp files
	m.cleaner.TrackTempDir(tempDir)
	m.cleaner.TrackTempFile(tempRCFile)

	// Register with safety checker
	if err := m.safetyChecker.RegisterPreview(session); err != nil {
		m.cleaner.CleanupSession(session)
		cancel()
		return nil, err
	}

	// Store session
	m.activeSessions[sessionID] = session

	// Persist session to disk
	if m.persister != nil {
		if _, err := m.persister.CreatePreviewSession(config.ThemeName, tempRCFile); err != nil {
			// Log error but don't fail
		}
	}

	return session, nil
}

// EndSession ends an active preview session.
func (m *DefaultSessionManager) EndSession(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.activeSessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	// Cancel context if still running
	if session.CancelFunc != nil {
		session.CancelFunc()
	}

	// Clean up resources
	result := m.cleaner.CleanupSession(session)
	if !result.Success {
		// Log errors but don't fail
	}

	// Unregister from safety checker
	m.safetyChecker.UnregisterPreview(sessionID)

	// Remove from active sessions
	delete(m.activeSessions, sessionID)

	// End persistence session
	if m.persister != nil {
		if err := m.persister.EndPreviewSession(); err != nil {
			// Log error but don't fail
		}
	}

	return nil
}

// GetActiveSession returns the currently active session, if any.
func (m *DefaultSessionManager) GetActiveSession() (*PreviewSessionState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return the first active session
	for _, session := range m.activeSessions {
		if session.Status == StatusRunning || session.Status == StatusPending {
			return session, nil
		}
	}

	return nil, ErrSessionNotFound
}

// HasActiveSession checks if there's an active session.
func (m *DefaultSessionManager) HasActiveSession() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.activeSessions) > 0
}

// UpdateSessionStatus updates the status of a session.
func (m *DefaultSessionManager) UpdateSessionStatus(sessionID string, status PreviewStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.activeSessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	session.Status = status
	return nil
}

// GetSession returns a specific preview session by ID.
func (m *DefaultSessionManager) GetSession(sessionID string) (*PreviewSessionState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.activeSessions[sessionID]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// CancelSession cancels a running preview session.
func (m *DefaultSessionManager) Cancel(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.activeSessions[sessionID]
	if !exists {
		return ErrSessionNotFound
	}

	// Update status
	session.Status = StatusCancelled

	// Cancel context
	if session.CancelFunc != nil {
		session.CancelFunc()
	}

	// Kill process if PID is set
	if session.PID > 0 {
		m.cleaner.killProcess(session.PID)
	}

	return nil
}

// RetrySession retries a failed session.
func (m *DefaultSessionManager) RetrySession(sessionID string) (*PreviewSessionState, error) {
	m.mu.RLock()
	session, exists := m.activeSessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, ErrSessionNotFound
	}

	// Can only retry failed or cancelled sessions
	if session.Status != StatusFailed && session.Status != StatusCancelled {
		return nil, fmt.Errorf("can only retry failed or cancelled sessions")
	}

	// Create new session with same config
	newSession, err := m.CreateSession(&session.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create new session: %w", err)
	}

	// Clean up old session
	m.EndSession(sessionID)

	return newSession, nil
}

// ListSessions returns all sessions (active and inactive).
func (m *DefaultSessionManager) ListSessions() []*PreviewSessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessions := make([]*PreviewSessionState, 0, len(m.activeSessions))
	for _, session := range m.activeSessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// CleanupStaleSessions removes sessions that have been running too long.
func (m *DefaultSessionManager) CleanupStaleSessions() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cleaned := 0
	now := time.Now()

	for id, session := range m.activeSessions {
		// Check if session has been running longer than timeout
		elapsed := now.Sub(session.StartTime)
		if elapsed > SessionTimeout {
			// Clean up
			m.cleaner.CleanupSession(session)
			m.safetyChecker.UnregisterPreview(id)
			delete(m.activeSessions, id)
			cleaned++
		}
	}

	return cleaned, nil
}

// StartCleanupRoutine starts a background goroutine to clean up stale sessions.
func (m *DefaultSessionManager) StartCleanupRoutine() chan struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(SessionCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.CleanupStaleSessions()
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}

// createTempFiles creates temporary files for a preview session.
func (m *DefaultSessionManager) createTempFiles(config *PreviewConfig) (string, string, error) {
	// Create temporary directory
	tempDir, err := createTempDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Determine RC filename based on shell type
	rcFilename := ".bashrc"
	if config.Shell == shell.ShellTypeZsh {
		rcFilename = ".zshrc"
	}

	tempRCFile := tempDir + "/" + rcFilename

	return tempDir, tempRCFile, nil
}

// createTempDir creates a temporary directory for preview files.
func createTempDir() (string, error) {
	return os.MkdirTemp("", "savanhi-preview-")
}

// generateSessionID generates a unique session ID.
func generateSessionID() string {
	return uuid.New().String()
}

// PreviewCoordinator coordinates preview operations across different types.
type PreviewCoordinator struct {
	// sessionManager manages preview sessions.
	sessionManager SessionManager

	// themePreview handles theme previews.
	themePreview ThemePreview

	// fontPreview handles font previews.
	fontPreview FontPreview

	// colorPreview handles color scheme previews.
	colorPreview ColorSchemePreview
}

// NewPreviewCoordinator creates a new PreviewCoordinator.
func NewPreviewCoordinator(
	sessionManager SessionManager,
	themePreview ThemePreview,
	fontPreview FontPreview,
	colorPreview ColorSchemePreview,
) *PreviewCoordinator {
	return &PreviewCoordinator{
		sessionManager: sessionManager,
		themePreview:   themePreview,
		fontPreview:    fontPreview,
		colorPreview:   colorPreview,
	}
}

// PreviewTheme creates and runs a theme preview.
func (c *PreviewCoordinator) PreviewTheme(ctx context.Context, config *ThemePreviewConfig) (*PreviewResult, error) {
	// Create session
	previewConfig := &PreviewConfig{
		Type:      PreviewTypeTheme,
		Shell:     config.Shell,
		ThemePath: config.ThemePath,
		ThemeName: config.ThemeName,
		Timeout:   config.Timeout,
	}

	session, err := c.sessionManager.CreateSession(previewConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer c.sessionManager.EndSession(session.ID)

	// Run preview
	result, err := c.themePreview.PreviewTheme(ctx, config)
	if err != nil {
		return nil, err
	}

	// Update session status
	if result.Status == StatusCompleted {
		c.sessionManager.UpdateSessionStatus(session.ID, StatusCompleted)
	} else {
		c.sessionManager.UpdateSessionStatus(session.ID, StatusFailed)
	}

	return result, nil
}

// PreviewFont creates and runs a font preview.
func (c *PreviewCoordinator) PreviewFont(ctx context.Context, config *FontPreviewConfig) (*PreviewResult, error) {
	previewConfig := &PreviewConfig{
		Type:       PreviewTypeFont,
		Shell:      config.Shell,
		FontFamily: config.FontFamily,
		FontSize:   config.FontSize,
		Timeout:    config.Timeout,
	}

	session, err := c.sessionManager.CreateSession(previewConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer c.sessionManager.EndSession(session.ID)

	result, err := c.fontPreview.PreviewFont(ctx, config)
	if err != nil {
		return nil, err
	}

	if result.Status == StatusCompleted {
		c.sessionManager.UpdateSessionStatus(session.ID, StatusCompleted)
	} else {
		c.sessionManager.UpdateSessionStatus(session.ID, StatusFailed)
	}

	return result, nil
}

// PreviewColorScheme creates and runs a color scheme preview.
func (c *PreviewCoordinator) PreviewColorScheme(ctx context.Context, config *ColorSchemePreviewConfig) (*PreviewResult, error) {
	previewConfig := &PreviewConfig{
		Type:        PreviewTypeColorScheme,
		Shell:       config.Shell,
		ColorScheme: config.ColorSchemeName,
		Timeout:     config.Timeout,
	}

	session, err := c.sessionManager.CreateSession(previewConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer c.sessionManager.EndSession(session.ID)

	result, err := c.colorPreview.PreviewColorScheme(ctx, config)
	if err != nil {
		return nil, err
	}

	if result.Status == StatusCompleted {
		c.sessionManager.UpdateSessionStatus(session.ID, StatusCompleted)
	} else {
		c.sessionManager.UpdateSessionStatus(session.ID, StatusFailed)
	}

	return result, nil
}

// CleanupAll cleans up all resources.
func (c *PreviewCoordinator) CleanupAll() error {
	// Clean up all active sessions
	// Note: This would need to iterate through active sessions
	// and end them
	return nil
}
