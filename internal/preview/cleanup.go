// Package preview provides live preview capabilities for Savanhi Shell.
// This file implements cleanup and safety mechanisms for previews.
package preview

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/savanhi/shell/pkg/shell"
)

// Cleanup and safety constants.
const (
	// DefaultProcessKillTimeout is the time to wait before SIGKILL.
	DefaultProcessKillTimeout = 2 * time.Second

	// DefaultGracefulShutdownTimeout is the time to wait for graceful shutdown.
	DefaultGracefulShutdownTimeout = 100 * time.Millisecond

	// MaxPreviewTimeout is the maximum allowed preview timeout.
	MaxPreviewTimeout = 60 * time.Second

	// MinPreviewTimeout is the minimum allowed preview timeout.
	MinPreviewTimeout = 1 * time.Second
)

// Common cleanup errors.
var (
	// ErrCleanupFailed indicates cleanup failed.
	ErrCleanupFailed = errors.New("cleanup failed")

	// ErrProcessKillFailed indicates failed to kill process.
	ErrProcessKillFailed = errors.New("failed to kill process")

	// ErrTempFileRemovalFailed indicates failed to remove temp files.
	ErrTempFileRemovalFailed = errors.New("failed to remove temp files")

	// ErrTimeoutTooHigh indicates timeout exceeds maximum.
	ErrTimeoutTooHigh = errors.New("timeout exceeds maximum allowed")

	// ErrTimeoutTooLow indicates timeout below minimum.
	ErrTimeoutTooLow = errors.New("timeout below minimum allowed")

	// ErrInvalidConfig indicates invalid configuration.
	ErrInvalidConfig = errors.New("invalid preview configuration")
)

// CleanupOptions contains options for cleanup operations.
type CleanupOptions struct {
	// RemoveTempFiles indicates whether to remove temporary files.
	RemoveTempFiles bool

	// KillOrphanProcesses indicates whether to kill orphan processes.
	KillOrphanProcesses bool

	// GracefulTimeout is the timeout for graceful shutdown.
	GracefulTimeout time.Duration

	// ForceKillAfter is the time to wait before force killing.
	ForceKillAfter time.Duration
}

// DefaultCleanupOptions returns default cleanup options.
func DefaultCleanupOptions() CleanupOptions {
	return CleanupOptions{
		RemoveTempFiles:     true,
		KillOrphanProcesses: true,
		GracefulTimeout:     DefaultGracefulShutdownTimeout,
		ForceKillAfter:      DefaultProcessKillTimeout,
	}
}

// PreviewSafetyChecker provides safety validation for previews.
type PreviewSafetyChecker struct {
	// mu protects concurrent access.
	mu sync.Mutex

	// activePreviews tracks active preview sessions.
	activePreviews map[string]*PreviewSessionState

	// maxConcurrentPreviews is the maximum allowed concurrent previews.
	maxConcurrentPreviews int
}

// NewPreviewSafetyChecker creates a new PreviewSafetyChecker.
func NewPreviewSafetyChecker() *PreviewSafetyChecker {
	return &PreviewSafetyChecker{
		activePreviews:        make(map[string]*PreviewSessionState),
		maxConcurrentPreviews: 3, // Limit to 3 concurrent previews
	}
}

// ValidateConfig validates a preview configuration.
func (p *PreviewSafetyChecker) ValidateConfig(config *PreviewConfig) error {
	if config == nil {
		return fmt.Errorf("%w: config is nil", ErrInvalidConfig)
	}

	// Validate timeout
	if config.Timeout > 0 {
		if config.Timeout > MaxPreviewTimeout {
			return fmt.Errorf("%w: timeout %v exceeds maximum %v",
				ErrTimeoutTooHigh, config.Timeout, MaxPreviewTimeout)
		}
		if config.Timeout < MinPreviewTimeout {
			return fmt.Errorf("%w: timeout %v below minimum %v",
				ErrTimeoutTooLow, config.Timeout, MinPreviewTimeout)
		}
	}

	// Validate shell type
	validShells := map[shell.ShellType]bool{
		shell.ShellTypeBash: true,
		shell.ShellTypeZsh:  true,
	}
	if !validShells[config.Shell] {
		return fmt.Errorf("%w: unsupported shell type %s",
			ErrInvalidConfig, config.Shell)
	}

	// Validate theme path if provided
	if config.ThemePath != "" {
		if !fileExists(config.ThemePath) {
			// Allow non-existent paths for bundled themes
			// The previewer will handle this
		}
	}

	// Validate environment variables for special characters
	for key, value := range config.Environment {
		if err := validateEnvVar(key, value); err != nil {
			return fmt.Errorf("%w: invalid environment variable %s: %v",
				ErrInvalidConfig, key, err)
		}
	}

	return nil
}

// EnforceTimeout ensures the preview does not exceed the timeout.
func (p *PreviewSafetyChecker) EnforceTimeout(parentCtx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// Cap timeout to maximum
	if timeout > MaxPreviewTimeout {
		timeout = MaxPreviewTimeout
	}
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	return context.WithTimeout(parentCtx, timeout)
}

// RecoverPanic recovers from panics in preview goroutines.
// Should be called with defer in all preview goroutines.
func (p *PreviewSafetyChecker) RecoverPanic() func() {
	return func() {
		if r := recover(); r != nil {
			// Log the panic (would use proper logging in production)
			fmt.Fprintf(os.Stderr, "Preview panic recovered: %v\n", r)
		}
	}
}

// RegisterPreview registers an active preview session.
func (p *PreviewSafetyChecker) RegisterPreview(session *PreviewSessionState) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check concurrent limit
	if len(p.activePreviews) >= p.maxConcurrentPreviews {
		return fmt.Errorf("maximum concurrent previews (%d) reached",
			p.maxConcurrentPreviews)
	}

	p.activePreviews[session.ID] = session
	return nil
}

// UnregisterPreview unregisters a preview session.
func (p *PreviewSafetyChecker) UnregisterPreview(sessionID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.activePreviews, sessionID)
}

// GetActivePreviews returns all active preview sessions.
func (p *PreviewSafetyChecker) GetActivePreviews() []*PreviewSessionState {
	p.mu.Lock()
	defer p.mu.Unlock()

	previews := make([]*PreviewSessionState, 0, len(p.activePreviews))
	for _, session := range p.activePreviews {
		previews = append(previews, session)
	}
	return previews
}

// PreviewCleaner handles cleanup of preview resources.
type PreviewCleaner struct {
	// mu protects concurrent access.
	mu sync.Mutex

	// tempFiles tracks temporary files created.
	tempFiles map[string]bool

	// tempDirs tracks temporary directories created.
	tempDirs map[string]bool

	// processes tracks spawned processes.
	processes map[int]bool
}

// NewPreviewCleaner creates a new PreviewCleaner.
func NewPreviewCleaner() *PreviewCleaner {
	return &PreviewCleaner{
		tempFiles: make(map[string]bool),
		tempDirs:  make(map[string]bool),
		processes: make(map[int]bool),
	}
}

// TrackTempFile tracks a temporary file for cleanup.
func (c *PreviewCleaner) TrackTempFile(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tempFiles[path] = true
}

// TrackTempDir tracks a temporary directory for cleanup.
func (c *PreviewCleaner) TrackTempDir(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tempDirs[path] = true
}

// TrackProcess tracks a spawned process for cleanup.
func (c *PreviewCleaner) TrackProcess(pid int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.processes[pid] = true
}

// CleanupAll cleans up all tracked resources.
func (c *PreviewCleaner) CleanupAll() *CleanupResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := &CleanupResult{
		Success: true,
	}

	// Kill processes
	for pid := range c.processes {
		if err := c.killProcess(pid); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to kill process %d: %v", pid, err))
			result.Success = false
		} else {
			result.KilledProcesses = append(result.KilledProcesses, pid)
		}
	}
	c.processes = make(map[int]bool)

	// Remove temp files
	for path := range c.tempFiles {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to remove file %s: %v", path, err))
			result.Success = false
		} else {
			result.RemovedFiles = append(result.RemovedFiles, path)
		}
	}
	c.tempFiles = make(map[string]bool)

	// Remove temp directories
	for path := range c.tempDirs {
		if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to remove directory %s: %v", path, err))
			result.Success = false
		}
	}
	c.tempDirs = make(map[string]bool)

	return result
}

// CleanupSession cleans up resources for a specific session.
func (c *PreviewCleaner) CleanupSession(session *PreviewSessionState) *CleanupResult {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := &CleanupResult{
		Success: true,
	}

	// Kill process if running
	if session.PID > 0 {
		if err := c.killProcess(session.PID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to kill process %d: %v", session.PID, err))
			result.Success = false
		} else {
			result.KilledProcesses = append(result.KilledProcesses, session.PID)
		}
		delete(c.processes, session.PID)
	}

	// Remove temp RC file
	if session.TempRCFile != "" {
		if err := os.Remove(session.TempRCFile); err != nil && !os.IsNotExist(err) {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to remove RC file %s: %v", session.TempRCFile, err))
			result.Success = false
		} else {
			result.RemovedFiles = append(result.RemovedFiles, session.TempRCFile)
		}
		delete(c.tempFiles, session.TempRCFile)
	}

	// Remove temp directory
	if session.TempDir != "" {
		if err := os.RemoveAll(session.TempDir); err != nil && !os.IsNotExist(err) {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to remove temp dir %s: %v", session.TempDir, err))
			result.Success = false
		}
		delete(c.tempDirs, session.TempDir)
	}

	return result
}

// killProcess kills a process by PID.
func (c *PreviewCleaner) killProcess(pid int) error {
	if pid <= 0 {
		return nil
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// Try graceful termination first
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process may already be dead
		if !errors.Is(err, os.ErrProcessDone) {
			return err
		}
		return nil
	}

	// Wait for graceful shutdown
	time.Sleep(DefaultGracefulShutdownTimeout)

	// Check if still running
	if err := process.Signal(syscall.Signal(0)); err != nil {
		// Process is dead
		return nil
	}

	// Force kill
	if err := process.Kill(); err != nil {
		return err
	}

	return nil
}

// KillOrphanProcesses kills any orphaned preview processes.
// This is useful for cleanup on startup to handle previous crashes.
func KillOrphanProcesses() error {
	// Find processes owned by current user that might be orphaned previews
	// This is platform-specific
	switch runtime.GOOS {
	case "linux", "darwin":
		return killOrphanProcessesUnix()
	case "windows":
		return killOrphanProcessesWindows()
	default:
		return nil
	}
}

// killOrphanProcessesUnix kills orphan processes on Unix-like systems.
func killOrphanProcessesUnix() error {
	// Use ps to find processes with "savanhi-preview" in their name
	cmd := exec.Command("ps", "-u", os.Getenv("USER"), "-o", "pid,comm")
	output, err := cmd.Output()
	if err != nil {
		return nil // Don't fail on error
	}

	// Parse output and find orphan processes
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "savanhi-preview") {
			// Extract PID and kill
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				pid := 0
				if _, err := fmt.Sscanf(fields[0], "%d", &pid); err == nil && pid > 0 {
					// Kill the process
					process, _ := os.FindProcess(pid)
					if process != nil {
						process.Kill()
					}
				}
			}
		}
	}

	return nil
}

// killOrphanProcessesWindows kills orphan processes on Windows.
func killOrphanProcessesWindows() error {
	// Use tasklist/taskkill on Windows
	// This is a simplified implementation
	return nil
}

// validateEnvVar validates an environment variable key-value pair.
func validateEnvVar(key, value string) error {
	// Check for dangerous characters
	dangerousChars := []string{"\n", "\r", "\x00", "`", "$(", "${"}
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return fmt.Errorf("value contains dangerous character: %s", char)
		}
	}

	// Check for shell injection attempts
	if strings.Contains(value, ";") && strings.Contains(value, "&&") {
		return errors.New("potential command injection detected")
	}

	return nil
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// EnsureCleanup ensures cleanup is called even on panic.
func EnsureCleanup(cleaner *PreviewCleaner, session *PreviewSessionState) {
	if r := recover(); r != nil {
		// Clean up on panic
		cleaner.CleanupSession(session)
		panic(r) // Re-panic after cleanup
	}
}

// CleanupOnExit registers cleanup to run on program exit.
func CleanupOnExit(cleaner *PreviewCleaner) {
	// Register cleanup function
	runtime.SetFinalizer(cleaner, func(c *PreviewCleaner) {
		c.CleanupAll()
	})

	// Also register for interrupt handling
	// Note: In production, use signal.Notify for SIGINT/SIGTERM
}
