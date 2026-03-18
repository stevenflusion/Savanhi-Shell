// Package shell provides RC file manipulation for different shells.
// It supports marker-based section injection for safe RC file modifications.
package shell

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Marker constants for Savanhi sections.
const (
	MarkerStartPrefix = "# >>> savanhi-"
	MarkerStartSuffix = " >>>"
	MarkerEndPrefix   = "# <<< savanhi-"
	MarkerEndSuffix   = " <<<"
)

// Common errors.
var (
	// ErrRCNotFound indicates the RC file does not exist.
	ErrRCNotFound = errors.New("rc file not found")
	// ErrRCReadFailed indicates failed to read RC file.
	ErrRCReadFailed = errors.New("failed to read rc file")
	// ErrRCWriteFailed indicates failed to write RC file.
	ErrRCWriteFailed = errors.New("failed to write rc file")
	// ErrUnclosedMarker indicates a marker section was not properly closed.
	ErrUnclosedMarker = errors.New("unclosed marker section")
	// ErrDuplicateMarker indicates duplicate markers in RC file.
	ErrDuplicateMarker = errors.New("duplicate markers found")
	// ErrPermissionDenied indicates insufficient permissions.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrUnsupportedShell indicates the shell type is not supported.
	ErrUnsupportedShell = errors.New("unsupported shell type")
)

// ShellType represents the type of shell.
type ShellType string

const (
	ShellTypeBash ShellType = "bash"
	ShellTypeZsh  ShellType = "zsh"
	ShellTypeFish ShellType = "fish"
	ShellTypePwsh ShellType = "pwsh"
)

// Shell is the interface for shell RC file manipulation.
type Shell interface {
	// GetType returns the shell type.
	GetType() ShellType

	// GetName returns the shell name (zsh, bash, etc).
	GetName() string

	// GetRCPath returns the path to the RC file.
	GetRCPath() (string, error)

	// ReadRC reads the RC file content.
	ReadRC() (string, error)

	// WriteRC writes content to the RC file atomically.
	WriteRC(content string) error

	// InjectSection injects a marked section into the RC file.
	// The section is identified by the marker name.
	// Returns ErrDuplicateMarker if the section already exists.
	InjectSection(marker string, content string) error

	// RemoveSection removes a marked section from the RC file.
	// Returns nil if the section doesn't exist.
	RemoveSection(marker string) error

	// HasSection checks if a marked section exists.
	HasSection(marker string) (bool, error)

	// GetSection returns the content of a marked section.
	// Returns empty string if section doesn't exist.
	GetSection(marker string) (string, error)

	// EnsureRCFile ensures the RC file exists, creating it if necessary.
	EnsureRCFile() error

	// Backup creates a backup of the RC file.
	Backup() (string, error)

	// Restore restores the RC file from a backup.
	Restore(backupPath string) error
}

// BaseShell provides common functionality for all shell implementations.
type BaseShell struct {
	Type    ShellType
	Name    string
	RCFile  string
	HomeDir string
}

// GetType returns the shell type.
func (s *BaseShell) GetType() ShellType {
	return s.Type
}

// GetName returns the shell name.
func (s *BaseShell) GetName() string {
	return s.Name
}

// GetRCPath returns the path to the RC file.
func (s *BaseShell) GetRCPath() (string, error) {
	if s.RCFile != "" {
		return s.RCFile, nil
	}

	home, err := s.getHomeDir()
	if err != nil {
		return "", err
	}

	switch s.Type {
	case ShellTypeBash:
		return filepath.Join(home, ".bashrc"), nil
	case ShellTypeZsh:
		return filepath.Join(home, ".zshrc"), nil
	case ShellTypeFish:
		return filepath.Join(home, ".config", "fish", "config.fish"), nil
	case ShellTypePwsh:
		// PowerShell profile location varies by platform
		if runtime.GOOS == "windows" {
			return filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"), nil
		}
		return filepath.Join(home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1"), nil
	default:
		return "", ErrUnsupportedShell
	}
}

// getHomeDir returns the user's home directory.
func (s *BaseShell) getHomeDir() (string, error) {
	if s.HomeDir != "" {
		return s.HomeDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return home, nil
}

// ReadRC reads the RC file content.
func (s *BaseShell) ReadRC() (string, error) {
	rcPath, err := s.GetRCPath()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(rcPath)
	if os.IsNotExist(err) {
		return "", ErrRCNotFound
	}
	if err != nil {
		if os.IsPermission(err) {
			return "", ErrPermissionDenied
		}
		return "", fmt.Errorf("%w: %v", ErrRCReadFailed, err)
	}

	return string(content), nil
}

// WriteRC writes content to the RC file atomically.
func (s *BaseShell) WriteRC(content string) error {
	rcPath, err := s.GetRCPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	rcDir := filepath.Dir(rcPath)
	if err := os.MkdirAll(rcDir, 0755); err != nil {
		return fmt.Errorf("failed to create RC directory: %w", err)
	}

	// Write atomically (write to temp, then rename)
	tempPath := rcPath + ".savanhi-tmp"

	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return fmt.Errorf("%w: %v", ErrRCWriteFailed, err)
	}

	if err := os.Rename(tempPath, rcPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("%w: %v", ErrRCWriteFailed, err)
	}

	return nil
}

// EnsureRCFile ensures the RC file exists, creating it if necessary.
func (s *BaseShell) EnsureRCFile() error {
	rcPath, err := s.GetRCPath()
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(rcPath); os.IsNotExist(err) {
		// Create directory
		rcDir := filepath.Dir(rcPath)
		if err := os.MkdirAll(rcDir, 0755); err != nil {
			return fmt.Errorf("failed to create RC directory: %w", err)
		}

		// Create empty file
		if err := os.WriteFile(rcPath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create RC file: %w", err)
		}
	}

	return nil
}

// Backup creates a backup of the RC file.
func (s *BaseShell) Backup() (string, error) {
	rcPath, err := s.GetRCPath()
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(rcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrRCNotFound
		}
		return "", fmt.Errorf("failed to read RC file: %w", err)
	}

	// Create backup in same directory
	backupPath := rcPath + ".backup"
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	return backupPath, nil
}

// Restore restores the RC file from a backup.
func (s *BaseShell) Restore(backupPath string) error {
	content, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	return s.WriteRC(string(content))
}

// formatStartMarker formats the start marker for a section.
func formatStartMarker(marker string) string {
	return MarkerStartPrefix + marker + MarkerStartSuffix
}

// formatEndMarker formats the end marker for a section.
func formatEndMarker(marker string) string {
	return MarkerEndPrefix + marker + MarkerEndSuffix
}

// InjectSection injects a marked section into the RC content.
func (s *BaseShell) InjectSection(marker string, content string) error {
	rcContent, err := s.ReadRC()
	if err != nil && err != ErrRCNotFound {
		return err
	}

	// Check if section already exists
	if has, _ := s.HasSection(marker); has {
		// Remove existing section first
		rcContent, err = s.removeSectionFromContent(rcContent, marker)
		if err != nil {
			return err
		}
	}

	// Format the section with markers
	startMarker := formatStartMarker(marker)
	endMarker := formatEndMarker(marker)

	section := fmt.Sprintf("\n%s\n%s\n%s\n", startMarker, content, endMarker)

	// Append to content
	newContent := rcContent + section

	return s.WriteRC(newContent)
}

// RemoveSection removes a marked section from the RC content.
func (s *BaseShell) RemoveSection(marker string) error {
	rcContent, err := s.ReadRC()
	if err != nil {
		if err == ErrRCNotFound {
			return nil // No file to remove from
		}
		return err
	}

	newContent, err := s.removeSectionFromContent(rcContent, marker)
	if err != nil {
		return err
	}

	return s.WriteRC(newContent)
}

// removeSectionFromContent removes a marked section from content.
func (s *BaseShell) removeSectionFromContent(content string, marker string) (string, error) {
	startMarker := formatStartMarker(marker)
	endMarker := formatEndMarker(marker)

	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return content, nil // Section not found
	}

	endIdx := strings.Index(content, endMarker)
	if endIdx == -1 {
		return "", ErrUnclosedMarker
	}

	// Include the newline after the end marker
	endIdx += len(endMarker)
	for endIdx < len(content) && (content[endIdx] == '\n' || content[endIdx] == '\r') {
		endIdx++
	}

	// Remove section (including trailing newlines before it)
	newContent := content[:startIdx] + content[endIdx:]

	// Clean up leading newlines
	newContent = strings.TrimRight(newContent, "\n\r") + "\n"

	return newContent, nil
}

// HasSection checks if a marked section exists.
func (s *BaseShell) HasSection(marker string) (bool, error) {
	rcContent, err := s.ReadRC()
	if err != nil {
		if err == ErrRCNotFound {
			return false, nil
		}
		return false, err
	}

	startMarker := formatStartMarker(marker)
	return strings.Contains(rcContent, startMarker), nil
}

// GetSection returns the content of a marked section.
func (s *BaseShell) GetSection(marker string) (string, error) {
	rcContent, err := s.ReadRC()
	if err != nil {
		return "", err
	}

	startMarker := formatStartMarker(marker)
	endMarker := formatEndMarker(marker)

	startIdx := strings.Index(rcContent, startMarker)
	if startIdx == -1 {
		return "", nil // Section not found
	}

	endIdx := strings.Index(rcContent, endMarker)
	if endIdx == -1 {
		return "", ErrUnclosedMarker
	}

	// Extract content between markers
	content := rcContent[startIdx+len(startMarker) : endIdx]
	return strings.TrimSpace(content), nil
}

// DetectShellType detects the current shell type.
func DetectShellType() ShellType {
	shell := os.Getenv("SHELL")
	if shell == "" {
		// Fallback on Windows
		shell = os.Getenv("PSModulePath")
		if shell != "" {
			return ShellTypePwsh
		}
		return ShellTypeBash // Default fallback
	}

	switch {
	case strings.Contains(shell, "zsh"):
		return ShellTypeZsh
	case strings.Contains(shell, "bash"):
		return ShellTypeBash
	case strings.Contains(shell, "fish"):
		return ShellTypeFish
	case strings.Contains(shell, "pwsh") || strings.Contains(shell, "powershell"):
		return ShellTypePwsh
	default:
		return ShellTypeBash // Default fallback
	}
}

// DetectShellVersion detects the version of the installed shell.
func DetectShellVersion(shellType ShellType) string {
	var cmd *exec.Cmd

	switch shellType {
	case ShellTypeZsh:
		cmd = exec.Command("zsh", "--version")
	case ShellTypeBash:
		cmd = exec.Command("bash", "--version")
	case ShellTypeFish:
		cmd = exec.Command("fish", "--version")
	case ShellTypePwsh:
		cmd = exec.Command("pwsh", "--version")
	default:
		return "unknown"
	}

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(output))
}
