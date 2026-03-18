// Package preview provides live preview capabilities for Savanhi Shell.
// This file implements subshell spawning with timeout and cleanup.
package preview

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/savanhi/shell/pkg/shell"
)

// Common errors for subshell operations.
var (
	// ErrShellNotFound indicates the shell executable was not found.
	ErrShellNotFound = errors.New("shell executable not found")

	// ErrSubshellTimeout indicates the subshell exceeded its timeout.
	ErrSubshellTimeout = errors.New("subshell timeout exceeded")

	// ErrSubshellFailed indicates the subshell process failed to start.
	ErrSubshellFailed = errors.New("subshell process failed")

	// ErrContextCancelled indicates the context was cancelled.
	ErrContextCancelled = errors.New("context cancelled")

	// ErrNoShellPath indicates the shell path could not be determined.
	ErrNoShellPath = errors.New("could not determine shell path")

	// ErrTempFileCreation indicates failed to create temporary files.
	ErrTempFileCreation = errors.New("failed to create temporary files")

	// ErrProcessNotFound indicates the process was not found.
	ErrProcessNotFound = errors.New("process not found")
)

// DefaultSubsheller is the default implementation of SubshellSpawner.
type DefaultSubsheller struct {
	// mu protects the processes map.
	mu sync.Mutex

	// processes tracks all spawned processes by PID.
	processes map[int]*exec.Cmd

	// tempDirs tracks all created temporary directories.
	tempDirs map[string]bool

	// tempFiles tracks all created temporary files.
	tempFiles map[string]bool

	// homeDir is the user's home directory.
	homeDir string
}

// NewDefaultSubsheller creates a new DefaultSubsheller.
func NewDefaultSubsheller() (*DefaultSubsheller, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	return &DefaultSubsheller{
		processes: make(map[int]*exec.Cmd),
		tempDirs:  make(map[string]bool),
		tempFiles: make(map[string]bool),
		homeDir:   homeDir,
	}, nil
}

// Spawn creates a new subshell process with the given configuration.
func (s *DefaultSubsheller) Spawn(ctx context.Context, config *SubshellConfig) (*SubshellResult, error) {
	// Validate configuration
	if err := s.validateConfig(config); err != nil {
		return nil, err
	}

	// Set default timeout if not specified
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Determine shell path
	shellPath, err := s.getShellPath(config)
	if err != nil {
		return nil, err
	}

	// Create temporary directory and RC file
	tempDir, tempRCFile, err := s.createTempFiles(config)
	if err != nil {
		return nil, err
	}

	// Build environment
	env := s.buildEnvironment(config)

	// Build command
	cmd := s.buildCommand(ctx, shellPath, config, tempRCFile, env)

	// Track the process
	s.mu.Lock()
	s.tempDirs[tempDir] = true
	s.tempFiles[tempRCFile] = true
	s.mu.Unlock()

	// Set up output capture
	var stdoutBuf, stderrBuf strings.Builder
	if config.CaptureStdout {
		cmd.Stdout = &stdoutBuf
	}
	if config.CaptureStderr {
		cmd.Stderr = &stderrBuf
	}

	// Create result
	result := &SubshellResult{
		TempFilePaths: []string{tempRCFile},
	}

	// Start the command
	startTime := time.Now()
	if err := cmd.Start(); err != nil {
		s.cleanupTempFiles(tempDir, tempRCFile)
		return nil, fmt.Errorf("%w: %v", ErrSubshellFailed, err)
	}

	// Track process
	s.mu.Lock()
	s.processes[cmd.Process.Pid] = cmd
	result.PID = cmd.Process.Pid
	s.mu.Unlock()

	// Wait for completion with context handling
	done := make(chan error, 1)
	go func() {
		err := cmd.Wait()
		done <- err
	}()

	select {
	case <-ctx.Done():
		// Context cancelled or timed out
		s.killProcess(cmd.Process.Pid)

		// Cleanup temp files
		s.cleanupTempFiles(tempDir, tempRCFile)

		// Remove process from tracking
		s.mu.Lock()
		delete(s.processes, cmd.Process.Pid)
		s.mu.Unlock()

		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrSubshellTimeout
		}
		return nil, ErrContextCancelled

	case err := <-done:
		result.Duration = time.Since(startTime)

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitErr.ExitCode()
			}
			result.Stderr = stderrBuf.String()
		} else {
			result.ExitCode = 0
		}

		result.Stdout = stdoutBuf.String()
		result.Stderr = stderrBuf.String()

		// Cleanup
		s.mu.Lock()
		delete(s.processes, cmd.Process.Pid)
		s.mu.Unlock()

		s.cleanupTempFiles(tempDir, tempRCFile)

		return result, nil
	}
}

// Kill terminates a subshell process by PID.
func (s *DefaultSubsheller) Kill(pid int) error {
	s.mu.Lock()
	cmd, exists := s.processes[pid]
	s.mu.Unlock()

	if !exists {
		return ErrProcessNotFound
	}

	return s.killProcess(cmd.Process.Pid)
}

// KillAll terminates all subshell processes spawned by this spawner.
func (s *DefaultSubsheller) KillAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	for pid, cmd := range s.processes {
		if err := s.killProcess(cmd.Process.Pid); err != nil {
			errs = append(errs, fmt.Errorf("failed to kill process %d: %w", pid, err))
		}
		delete(s.processes, pid)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors killing processes: %v", errs)
	}

	return nil
}

// CleanupTempFiles cleans up all temporary files created by this spawner.
func (s *DefaultSubsheller) CleanupTempFiles() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	for tempFile := range s.tempFiles {
		if err := os.Remove(tempFile); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("failed to remove temp file %s: %w", tempFile, err))
		}
		delete(s.tempFiles, tempFile)
	}

	for tempDir := range s.tempDirs {
		if err := os.RemoveAll(tempDir); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("failed to remove temp dir %s: %w", tempDir, err))
		}
		delete(s.tempDirs, tempDir)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors cleaning up: %v", errs)
	}

	return nil
}

// validateConfig validates the subshell configuration.
func (s *DefaultSubsheller) validateConfig(config *SubshellConfig) error {
	if config.ShellType == "" {
		return fmt.Errorf("shell type is required")
	}

	// Validate shell type
	validShells := map[shell.ShellType]bool{
		shell.ShellTypeBash: true,
		shell.ShellTypeZsh:  true,
		// fish and pwsh support is limited
	}

	if !validShells[config.ShellType] {
		return fmt.Errorf("unsupported shell type: %s", config.ShellType)
	}

	return nil
}

// getShellPath determines the shell executable path.
func (s *DefaultSubsheller) getShellPath(config *SubshellConfig) (string, error) {
	if config.ShellPath != "" {
		// Verify path exists
		if _, err := exec.LookPath(config.ShellPath); err != nil {
			return "", fmt.Errorf("%w: %s", ErrShellNotFound, config.ShellPath)
		}
		return config.ShellPath, nil
	}

	// Try to find shell in PATH
	shellNames := map[shell.ShellType]string{
		shell.ShellTypeBash: "bash",
		shell.ShellTypeZsh:  "zsh",
		shell.ShellTypeFish: "fish",
		shell.ShellTypePwsh: "pwsh",
	}

	shellName, ok := shellNames[config.ShellType]
	if !ok {
		return "", ErrNoShellPath
	}

	path, err := exec.LookPath(shellName)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrShellNotFound, shellName)
	}

	return path, nil
}

// createTempFiles creates a temporary directory and RC file.
func (s *DefaultSubsheller) createTempFiles(config *SubshellConfig) (string, string, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "savanhi-preview-*")
	if err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrTempFileCreation, err)
	}

	// Create temporary RC file
	rcContent := config.RCContent
	if rcContent == "" {
		rcContent = s.generateMinimalRC(config)
	}

	// Determine RC filename based on shell type
	rcFilename := ".bashrc"
	if config.ShellType == shell.ShellTypeZsh {
		rcFilename = ".zshrc"
	}

	tempRCFile := filepath.Join(tempDir, rcFilename)
	if err := os.WriteFile(tempRCFile, []byte(rcContent), 0644); err != nil {
		os.RemoveAll(tempDir)
		return "", "", fmt.Errorf("%w: failed to write RC file: %v", ErrTempFileCreation, err)
	}

	return tempDir, tempRCFile, nil
}

// generateMinimalRC generates a minimal RC file for the preview.
func (s *DefaultSubsheller) generateMinimalRC(config *SubshellConfig) string {
	var sb strings.Builder

	sb.WriteString("# Savanhi Shell Preview - Minimal RC\n")
	sb.WriteString("# This is an isolated preview environment\n\n")

	// Source system RC if exists (for basic functions)
	switch config.ShellType {
	case shell.ShellTypeBash:
		sb.WriteString("# Source minimal bash functions\n")
		sb.WriteString("[ -f /etc/bash.bashrc ] && source /etc/bash.bashrc\n")

	case shell.ShellTypeZsh:
		sb.WriteString("# Source minimal zsh functions\n")
		sb.WriteString("[ -f /etc/zsh/zshenv ] && source /etc/zsh/zshenv\n")
	}

	// Add environment variables
	if config.Environment != nil {
		sb.WriteString("\n# Preview environment variables\n")
		for key, value := range config.Environment {
			// Escape special characters in value
			escaped := s.escapeShellValue(value)
			sb.WriteString(fmt.Sprintf("export %s='%s'\n", key, escaped))
		}
	}

	// Add preview command if specified
	if config.Command != "" {
		sb.WriteString("\n# Preview command\n")
		sb.WriteString(config.Command + "\n")
	}

	return sb.String()
}

// escapeShellValue escapes special characters for shell environment variables.
func (s *DefaultSubsheller) escapeShellValue(value string) string {
	// Escape single quotes and backslashes
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `'`, `'\''`)
	return value
}

// buildEnvironment builds the environment for the subshell.
func (s *DefaultSubsheller) buildEnvironment(config *SubshellConfig) []string {
	// Start with current environment
	env := os.Environ()

	// Add custom environment variables
	if config.Environment != nil {
		for key, value := range config.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return env
}

// buildCommand builds the exec.Cmd for the subshell.
func (s *DefaultSubsheller) buildCommand(ctx context.Context, shellPath string, config *SubshellConfig, tempRCFile string, env []string) *exec.Cmd {
	// Build arguments
	args := []string{}

	// Add login shell flag for proper environment setup
	if config.ShellType == shell.ShellTypeZsh {
		args = append(args, "-l")
	}

	// Add interactive flag
	if config.Command == "" {
		args = append(args, "-i")
	}

	// Add command if specified
	if config.Command != "" {
		args = append(args, "-c", config.Command)
	}

	// Create command
	cmd := exec.CommandContext(ctx, shellPath, args...)

	// Set environment
	cmd.Env = env

	// Set working directory
	if config.WorkingDir != "" {
		cmd.Dir = config.WorkingDir
	} else {
		cmd.Dir = s.homeDir
	}

	// Set up stdout/stderr capture using buffers
	// Output will be captured and read by the caller via cmd.Stdout and cmd.Stderr
	// Note: The actual reading happens in Spawn() which sets up the buffers

	// Set process group for proper termination
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	return cmd
}

// killProcess kills a process by PID with SIGTERM then SIGKILL.
func (s *DefaultSubsheller) killProcess(pid int) error {
	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process may already be dead
		if !errors.Is(err, os.ErrProcessDone) {
			return err
		}
	}

	// Wait briefly for graceful shutdown
	time.Sleep(100 * time.Millisecond)

	// Check if still running
	if err := process.Signal(syscall.Signal(0)); err == nil {
		// Still running, send SIGKILL
		if err := process.Kill(); err != nil {
			return err
		}
	}

	return nil
}

// cleanupTempFiles removes temporary files and directories.
func (s *DefaultSubsheller) cleanupTempFiles(tempDir string, tempRCFile string) {
	// Remove temp RC file
	if tempRCFile != "" {
		os.Remove(tempRCFile)
	}

	// Remove temp directory
	if tempDir != "" {
		os.RemoveAll(tempDir)
	}

	// Update tracking
	s.mu.Lock()
	delete(s.tempFiles, tempRCFile)
	delete(s.tempDirs, tempDir)
	s.mu.Unlock()
}

// GetSubshellPath returns the path to the shell executable for the given type.
func GetSubshellPath(shellType shell.ShellType) (string, error) {
	shellNames := map[shell.ShellType]string{
		shell.ShellTypeBash: "bash",
		shell.ShellTypeZsh:  "zsh",
		shell.ShellTypeFish: "fish",
		shell.ShellTypePwsh: "pwsh",
	}

	shellName, ok := shellNames[shellType]
	if !ok {
		return "", ErrNoShellPath
	}

	path, err := exec.LookPath(shellName)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrShellNotFound, shellName)
	}

	return path, nil
}

// IsShellAvailable checks if a shell is available on the system.
func IsShellAvailable(shellType shell.ShellType) bool {
	_, err := GetSubshellPath(shellType)
	return err == nil
}

// GetDefaultShell returns the default shell for previews.
// Prefers zsh over bash if available.
func GetDefaultShell() shell.ShellType {
	// Prefer zsh if available
	if IsShellAvailable(shell.ShellTypeZsh) {
		return shell.ShellTypeZsh
	}
	// Fall back to bash
	if IsShellAvailable(shell.ShellTypeBash) {
		return shell.ShellTypeBash
	}
	// Default to bash
	return shell.ShellTypeBash
}

// DetectShellFromEnv detects the user's default shell from environment.
func DetectShellFromEnv() shell.ShellType {
	shellEnv := os.Getenv("SHELL")
	if shellEnv == "" {
		return GetDefaultShell()
	}

	switch {
	case strings.Contains(shellEnv, "zsh"):
		return shell.ShellTypeZsh
	case strings.Contains(shellEnv, "bash"):
		return shell.ShellTypeBash
	case strings.Contains(shellEnv, "fish"):
		return shell.ShellTypeFish
	case strings.Contains(shellEnv, "pwsh") || strings.Contains(shellEnv, "powershell"):
		return shell.ShellTypePwsh
	default:
		return GetDefaultShell()
	}
}

// Platform detection for shell paths
func init() {
	// Ensure we're on a supported platform
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		// Windows and other platforms have limited support
	}
}
