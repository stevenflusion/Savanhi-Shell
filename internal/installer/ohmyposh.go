// Package installer provides dependency installation and management.
// This file implements oh-my-posh specific installation.
package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// OhMyPoshInstaller handles oh-my-posh specific installation.
type OhMyPoshInstaller struct {
	// context is the installation context.
	context *InstallContext

	// binDir is the binary installation directory.
	binDir string
}

// NewOhMyPoshInstaller creates a new oh-my-posh installer.
func NewOhMyPoshInstaller(ctx *InstallContext) *OhMyPoshInstaller {
	binDir := ctx.BinDir
	if binDir == "" {
		homeDir, _ := os.UserHomeDir()
		binDir = filepath.Join(homeDir, ".local", "bin")
	}

	return &OhMyPoshInstaller{
		context: ctx,
		binDir:  binDir,
	}
}

// Install installs oh-my-posh.
func (i *OhMyPoshInstaller) Install(ctx context.Context, opts *Options) (*InstallResult, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	result := &InstallResult{
		Component: "oh-my-posh",
	}

	// Check if already installed
	if i.IsInstalled() && !i.context.Force {
		version, _ := i.GetVersion()
		result.Success = true
		result.Version = version
		result.InstalledPath, _ = i.GetPath()
		return result, nil
	}

	// Ensure bin directory exists
	if err := os.MkdirAll(i.binDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Download oh-my-posh
	downloadURL := i.getDownloadURL()

	installer := &DefaultInstaller{context: i.context}
	downloadResult, err := installer.downloadFile(ctx, downloadURL, i.binDir, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to download oh-my-posh: %w", err)
	}

	// Determine target path
	targetPath := filepath.Join(i.binDir, "oh-my-posh")
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	// Move to final location
	if downloadResult.LocalPath != targetPath {
		if err := os.Rename(downloadResult.LocalPath, targetPath); err != nil {
			// Try copy if rename fails
			if err := copyFile(downloadResult.LocalPath, targetPath); err != nil {
				return nil, fmt.Errorf("failed to install oh-my-posh: %w", err)
			}
			os.Remove(downloadResult.LocalPath)
		}
	}

	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to make oh-my-posh executable: %w", err)
	}

	// Verify installation
	version, err := i.GetVersion()
	if err != nil {
		return nil, fmt.Errorf("oh-my-posh installed but version check failed: %w", err)
	}

	// Add to PATH if needed
	if err := i.ensureInPath(); err != nil {
		result.Warnings = append(result.Warnings, err.Error())
	}

	result.Success = true
	result.Version = version
	result.InstalledPath = targetPath
	result.RequiresRestart = true

	return result, nil
}

// IsInstalled checks if oh-my-posh is installed.
func (i *OhMyPoshInstaller) IsInstalled() bool {
	_, err := exec.LookPath("oh-my-posh")
	if err == nil {
		return true
	}

	// Check bin directory
	targetPath := filepath.Join(i.binDir, "oh-my-posh")
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	if _, err := os.Stat(targetPath); err == nil {
		return true
	}

	return false
}

// GetVersion returns the installed version.
func (i *OhMyPoshInstaller) GetVersion() (string, error) {
	path, err := i.GetPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(path, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	// Parse version from output (e.g., "oh-my-posh version 19.0.0")
	version := strings.TrimSpace(string(output))
	if strings.HasPrefix(version, "oh-my-posh") {
		parts := strings.Fields(version)
		if len(parts) >= 3 {
			return parts[2], nil
		}
	}

	return version, nil
}

// GetPath returns the path to oh-my-posh binary.
func (i *OhMyPoshInstaller) GetPath() (string, error) {
	// Check PATH first
	if path, err := exec.LookPath("oh-my-posh"); err == nil {
		return path, nil
	}

	// Check bin directory
	targetPath := filepath.Join(i.binDir, "oh-my-posh")
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	if _, err := os.Stat(targetPath); err == nil {
		return targetPath, nil
	}

	return "", fmt.Errorf("oh-my-posh not found")
}

// getDownloadURL returns the download URL for the current platform.
func (i *OhMyPoshInstaller) getDownloadURL() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map architecture
	switch arch {
	case "amd64":
		arch = "amd64"
	case "arm64":
		// oh-my-posh uses arm64 for Apple Silicon and arm64 for Linux ARM
		arch = "arm64"
	}

	// Map OS
	switch osName {
	case "darwin":
		osName = "darwin"
	case "linux":
		osName = "linux"
	case "windows":
		osName = "windows"
	}

	// oh-my-posh binary naming: posh-{os}-{arch}{.exe}
	filename := fmt.Sprintf("posh-%s-%s", osName, arch)
	if osName == "windows" {
		filename += ".exe"
	}

	return fmt.Sprintf("https://github.com/jandedobbeleer/oh-my-posh/releases/latest/download/%s", filename)
}

// ensureInPath ensures oh-my-posh is in PATH.
func (i *OhMyPoshInstaller) ensureInPath() error {
	// Check if already in PATH
	if _, err := exec.LookPath("oh-my-posh"); err == nil {
		return nil
	}

	// binDir should be in PATH for future shells
	// This is handled by RC file modification
	return nil
}

// Uninstall removes oh-my-posh.
func (i *OhMyPoshInstaller) Uninstall() error {
	targetPath := filepath.Join(i.binDir, "oh-my-posh")
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove oh-my-posh: %w", err)
	}

	return nil
}

// InstallThemes installs oh-my-posh themes to the config directory.
func (i *OhMyPoshInstaller) InstallThemes(ctx context.Context, themeDir string, opts *Options) error {
	// Theme installation is optional and handled by the theme manager
	// This is here for completeness but not used in MVP
	return nil
}
