// Package detector provides system detection capabilities.
// This file implements shell detection for zsh, bash, fish, and other shells.
package detector

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// shellDetector implements ShellDetector interface.
type shellDetector struct{}

// NewShellDetector creates a new shell detector.
func NewShellDetector() ShellDetector {
	return &shellDetector{}
}

// Detect implements ShellDetector.Detect.
func (d *shellDetector) Detect() (*ShellInfo, error) {
	info := &ShellInfo{
		IsDefault: false,
	}

	// Get the current shell from SHELL environment variable
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		// Fallback to detecting from other methods
		shellPath = "/bin/bash" // Default fallback
	}

	info.Path = shellPath

	// Get available shells from /etc/shells
	info.AvailableShells = d.parseEtcShells()

	// Determine shell name from path
	info.Name = d.getShellTypeFromPath(shellPath)

	// Get shell version
	info.Version = d.getShellVersion(string(info.Name))

	// Determine RC file path
	info.RCFile = d.getRCFilePath(info.Name)
	info.RCFileExists = d.fileExists(info.RCFile)

	// Determine config directory
	info.ConfigDir = d.getConfigDir(info.Name)

	// Check if this is the default shell
	info.IsDefault = d.isDefaultShell(shellPath)

	return info, nil
}

// parseEtcShells reads and parses /etc/shells.
func (d *shellDetector) parseEtcShells() []string {
	file, err := os.Open("/etc/shells")
	if err != nil {
		return nil
	}
	defer file.Close()

	var shells []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Validate line is an absolute path
		if strings.HasPrefix(line, "/") {
			shells = append(shells, line)
		}
	}

	return shells
}

// getShellTypeFromPath determines the shell type from its path.
func (d *shellDetector) getShellTypeFromPath(path string) ShellType {
	basename := filepath.Base(path)
	switch basename {
	case "zsh":
		return ShellTypeZsh
	case "bash":
		return ShellTypeBash
	case "fish":
		return ShellTypeFish
	case "pwsh", "powershell":
		return ShellTypePwsh
	default:
		return ShellTypeUnknown
	}
}

// getShellVersion returns the shell version.
func (d *shellDetector) getShellVersion(shellName string) string {
	var cmd *exec.Cmd
	switch shellName {
	case "zsh":
		cmd = exec.Command("zsh", "--version")
	case "bash":
		cmd = exec.Command("bash", "--version")
	case "fish":
		cmd = exec.Command("fish", "--version")
	case "pwsh":
		cmd = exec.Command("pwsh", "--version")
	default:
		return "unknown"
	}

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse version from output
	parts := strings.Fields(string(output))
	if len(parts) >= 2 {
		return parts[1]
	}

	return strings.TrimSpace(string(output))
}

// getRCFilePath returns the path to the shell's RC file.
func (d *shellDetector) getRCFilePath(shellName ShellType) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	switch shellName {
	case ShellTypeZsh:
		return filepath.Join(homeDir, ".zshrc")
	case ShellTypeBash:
		return filepath.Join(homeDir, ".bashrc")
	case ShellTypeFish:
		return filepath.Join(homeDir, ".config", "fish", "config.fish")
	case ShellTypePwsh:
		// PowerShell profile location varies by platform
		return filepath.Join(homeDir, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	default:
		return ""
	}
}

// getConfigDir returns the path to the shell's config directory.
func (d *shellDetector) getConfigDir(shellName ShellType) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configDir := filepath.Join(homeDir, ".config")

	switch shellName {
	case ShellTypeZsh:
		return filepath.Join(configDir, "zsh")
	case ShellTypeBash:
		return filepath.Join(configDir, "bash")
	case ShellTypeFish:
		return filepath.Join(configDir, "fish")
	case ShellTypePwsh:
		return filepath.Join(configDir, "powershell")
	default:
		return configDir
	}
}

// fileExists checks if a file exists.
func (d *shellDetector) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isDefaultShell checks if the given shell path is the default shell.
func (d *shellDetector) isDefaultShell(shellPath string) bool {
	// Compare with SHELL environment variable
	currentShell := os.Getenv("SHELL")
	return currentShell == shellPath
}
