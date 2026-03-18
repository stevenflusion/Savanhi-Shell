// Package installer provides dependency installation and management for Savanhi Shell.
// It handles downloading, installing, verifying, and rolling back components.
package installer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/savanhi/shell/internal/detector"
	"github.com/savanhi/shell/pkg/shell"
)

// Common errors returned by the installer package.
var (
	// ErrAlreadyInstalled indicates the component is already installed.
	ErrAlreadyInstalled = errors.New("component already installed")
	// ErrNotInstalled indicates the component is not installed.
	ErrNotInstalled = errors.New("component not installed")
	// ErrDownloadFailed indicates a download failure.
	ErrDownloadFailed = errors.New("download failed")
	// ErrChecksumMismatch indicates checksum verification failed.
	ErrChecksumMismatch = errors.New("checksum mismatch")
	// ErrInstallFailed indicates an installation failure.
	ErrInstallFailed = errors.New("installation failed")
	// ErrVerifyFailed indicates verification failed.
	ErrVerifyFailed = errors.New("verification failed")
	// ErrRollbackFailed indicates a rollback failure.
	ErrRollbackFailed = errors.New("rollback failed")
	// ErrUnsupportedPlatform indicates the platform is not supported.
	ErrUnsupportedPlatform = errors.New("unsupported platform")
	// ErrDependencyNotMet indicates a dependency is not satisfied.
	ErrDependencyNotMet = errors.New("dependency not met")
	// ErrNetworkError indicates a network-related error.
	ErrNetworkError = errors.New("network error")
	// ErrPermissionDenied indicates insufficient permissions.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrInstallationCancelled indicates the installation was cancelled.
	ErrInstallationCancelled = errors.New("installation cancelled")
)

// Installer is the interface for installing components.
type Installer interface {
	// Install installs a component.
	Install(ctx context.Context, dep *Dependency, opts *Options) (*InstallResult, error)

	// Verify verifies a component is correctly installed.
	Verify(ctx context.Context, name string) (*VerificationResult, error)

	// Uninstall removes an installed component.
	Uninstall(ctx context.Context, name string) error

	// GetProgress returns a channel for progress updates.
	GetProgress() <-chan *InstallProgress

	// IsInstalled checks if a component is installed.
	IsInstalled(name string) bool

	// GetInstalledVersion returns the installed version of a component.
	GetInstalledVersion(name string) (string, error)
}

// DefaultInstaller is the default implementation of Installer.
type DefaultInstaller struct {
	// context is the installation context.
	context *InstallContext

	// resolver resolves dependencies.
	resolver *DependencyResolver

	// detector provides system detection.
	detector detector.Detector

	// shell is the shell interface for RC modifications.
	shell shell.Shell

	// progressChan is the channel for progress updates.
	progressChan chan *InstallProgress

	// installed tracks installed components.
	installed map[string]*InstallResult

	// downloadCache caches downloads.
	downloadCache map[string]string
}

// NewInstaller creates a new Installer with default settings.
func NewInstaller() (*DefaultInstaller, error) {
	ctx, err := NewInstallContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create install context: %w", err)
	}

	installer := &DefaultInstaller{
		context:       ctx,
		resolver:      NewDependencyResolver(),
		progressChan:  make(chan *InstallProgress, 100),
		installed:     make(map[string]*InstallResult),
		downloadCache: make(map[string]string),
	}

	return installer, nil
}

// NewInstallerWithDetector creates a new Installer with a custom detector.
func NewInstallerWithDetector(det detector.Detector) (*DefaultInstaller, error) {
	installer, err := NewInstaller()
	if err != nil {
		return nil, err
	}
	installer.detector = det
	return installer, nil
}

// NewInstallContext creates a default installation context.
func NewInstallContext() (*InstallContext, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Determine platform
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Determine binary directory
	binDir := filepath.Join(homeDir, ".local", "bin")
	if platform == "windows" {
		// On Windows, use AppData\Local\savanhi\bin
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			appData = filepath.Join(homeDir, "AppData", "Local")
		}
		binDir = filepath.Join(appData, "savanhi", "bin")
	}

	// Determine font directory based on platform
	var fontDir string
	switch platform {
	case "darwin":
		fontDir = filepath.Join(homeDir, "Library", "Fonts")
	default: // linux and others
		fontDir = filepath.Join(homeDir, ".local", "share", "fonts")
	}

	// Determine config directory
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		configDir = filepath.Join(homeDir, ".config")
	}
	configDir = filepath.Join(configDir, "savanhi")

	// Determine cache directory
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		cacheDir = filepath.Join(homeDir, ".cache")
	}
	cacheDir = filepath.Join(cacheDir, "savanhi")

	return &InstallContext{
		Context:    context.Background(),
		ConfigDir:  configDir,
		HomeDir:    homeDir,
		BinDir:     binDir,
		FontDir:    fontDir,
		CacheDir:   cacheDir,
		OS:         platform,
		Arch:       arch,
		PackageMgr: detectPackageManager(platform),
		Shell:      detectShell(),
	}, nil
}

// detectPackageManager detects the system package manager.
func detectPackageManager(platform string) string {
	if platform == "darwin" {
		// Check for Homebrew
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew"
		}
		// Check for MacPorts
		if _, err := exec.LookPath("port"); err == nil {
			return "port"
		}
		return "brew" // Default to Homebrew
	}

	// Linux package managers
	if _, err := exec.LookPath("apt"); err == nil {
		return "apt"
	}
	if _, err := exec.LookPath("apt-get"); err == nil {
		return "apt"
	}
	if _, err := exec.LookPath("dnf"); err == nil {
		return "dnf"
	}
	if _, err := exec.LookPath("yum"); err == nil {
		return "yum"
	}
	if _, err := exec.LookPath("pacman"); err == nil {
		return "pacman"
	}
	if _, err := exec.LookPath("zypper"); err == nil {
		return "zypper"
	}
	if _, err := exec.LookPath("apk"); err == nil {
		return "apk"
	}

	return "unknown"
}

// detectShell detects the current shell.
func detectShell() string {
	shellEnv := os.Getenv("SHELL")
	if shellEnv == "" {
		return "bash" // Default
	}

	if strings.Contains(shellEnv, "zsh") {
		return "zsh"
	}
	if strings.Contains(shellEnv, "bash") {
		return "bash"
	}
	if strings.Contains(shellEnv, "fish") {
		return "fish"
	}

	return "bash" // Default fallback
}

// Install installs a component.
func (i *DefaultInstaller) Install(ctx context.Context, dep *Dependency, opts *Options) (*InstallResult, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	startTime := time.Now()
	result := &InstallResult{
		Component: dep.Name,
	}

	// Report progress
	i.reportProgress(&InstallProgress{
		Component: dep.Name,
		Stage:     StageResolving,
		Percent:   0,
		Message:   "Resolving dependencies...",
	})

	// Check if already installed
	if dep.IsInstalled && !i.context.Force {
		i.reportProgress(&InstallProgress{
			Component: dep.Name,
			Stage:     StageCompleted,
			Percent:   100,
			Message:   "Already installed",
		})
		result.Success = true
		result.Version = dep.InstalledVersion
		result.InstalledPath = dep.InstallPath
		return result, nil
	}

	// Ensure installation directories exist
	if err := i.ensureDirectories(); err != nil {
		result.Error = err.Error()
		i.reportProgress(&InstallProgress{
			Component: dep.Name,
			Stage:     StageFailed,
			Error:     err.Error(),
		})
		return result, fmt.Errorf("failed to create directories: %w", err)
	}

	// Install based on type
	var err error
	switch dep.Type {
	case ComponentTypeBinary:
		err = i.installBinary(ctx, dep, opts, result)
	case ComponentTypeFont:
		err = i.installFont(ctx, dep, opts, result)
	case ComponentTypePackage:
		err = i.installPackage(ctx, dep, opts, result)
	default:
		err = fmt.Errorf("unsupported component type: %s", dep.Type)
	}

	if err != nil {
		result.Error = err.Error()
		i.reportProgress(&InstallProgress{
			Component: dep.Name,
			Stage:     StageFailed,
			Error:     err.Error(),
		})
		return result, err
	}

	// Run post-install commands
	if len(dep.PostInstallCommands) > 0 && !i.context.DryRun {
		if err := i.runPostInstallCommands(dep, result); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("post-install command failed: %v", err))
		}
	}

	// Verify installation if requested
	if !opts.SkipVerification && !i.context.DryRun {
		i.reportProgress(&InstallProgress{
			Component: dep.Name,
			Stage:     StageVerifyingInstall,
			Percent:   90,
			Message:   "Verifying installation...",
		})

		verifyResult, err := i.Verify(ctx, dep.Name)
		if err != nil || !verifyResult.Installed {
			result.Error = "verification failed"
			if err != nil {
				result.Error = err.Error()
			}
			i.reportProgress(&InstallProgress{
				Component: dep.Name,
				Stage:     StageFailed,
				Error:     result.Error,
			})
			return result, fmt.Errorf("%w: %s", ErrVerifyFailed, result.Error)
		}
	}

	// Mark as complete
	result.Success = true
	result.Duration = time.Since(startTime)

	i.reportProgress(&InstallProgress{
		Component: dep.Name,
		Stage:     StageCompleted,
		Percent:   100,
		Message:   "Installation complete",
	})

	// Store result
	i.installed[dep.Name] = result

	return result, nil
}

// Verify verifies a component is correctly installed.
func (i *DefaultInstaller) Verify(ctx context.Context, name string) (*VerificationResult, error) {
	result := &VerificationResult{
		Component: name,
		Checks:    []VerificationCheck{},
	}

	// Get component info
	dep := i.resolver.GetDependency(name)
	if dep == nil {
		return nil, fmt.Errorf("unknown component: %s", name)
	}

	// Check if binary exists in PATH
	if dep.Type == ComponentTypeBinary {
		check := VerificationCheck{
			Name:    "binary_in_path",
			Passed:  false,
			Message: "Binary not found in PATH",
		}

		path, err := exec.LookPath(name)
		if err == nil {
			check.Passed = true
			check.Message = fmt.Sprintf("Binary found at %s", path)
			result.Path = path
			result.InPATH = true
		}

		result.Checks = append(result.Checks, check)
	}

	// Run verify command if specified
	if dep.VerifyCommand != "" {
		check := VerificationCheck{
			Name:    "verify_command",
			Passed:  false,
			Message: "Verify command failed",
		}

		cmdParts := strings.Fields(dep.VerifyCommand)
		if len(cmdParts) > 0 {
			cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
			output, err := cmd.CombinedOutput()
			if err == nil {
				check.Passed = true
				check.Message = "Verify command passed"
				result.Working = true

				// Try to extract version from output
				if len(output) > 0 {
					outputStr := strings.TrimSpace(string(output))
					// Common version output formats: "version 1.2.3" or "v1.2.3" or just "1.2.3"
					lines := strings.Split(outputStr, "\n")
					if len(lines) > 0 {
						result.Version = extractVersion(lines[0])
					}
				}
			} else {
				check.Message = fmt.Sprintf("Command failed: %v", err)
			}
		}

		result.Checks = append(result.Checks, check)
	}

	// Determine overall installed status
	result.Installed = i.isComponentInstalled(dep, result)

	return result, nil
}

// Uninstall removes an installed component.
func (i *DefaultInstaller) Uninstall(ctx context.Context, name string) error {
	// Get component info
	dep := i.resolver.GetDependency(name)
	if dep == nil {
		return fmt.Errorf("unknown component: %s", name)
	}

	// Use uninstall command if specified
	if dep.UninstallCommand != "" {
		cmdParts := strings.Fields(dep.UninstallCommand)
		if len(cmdParts) > 0 {
			cmd := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("uninstall command failed: %w", err)
			}
			delete(i.installed, name)
			return nil
		}
	}

	// Remove based on type
	switch dep.Type {
	case ComponentTypeBinary:
		return i.uninstallBinary(dep)
	case ComponentTypeFont:
		return i.uninstallFont(dep)
	default:
		return fmt.Errorf("cannot uninstall component type: %s", dep.Type)
	}
}

// GetProgress returns a channel for progress updates.
func (i *DefaultInstaller) GetProgress() <-chan *InstallProgress {
	return i.progressChan
}

// IsInstalled checks if a component is installed.
func (i *DefaultInstaller) IsInstalled(name string) bool {
	if result, ok := i.installed[name]; ok {
		return result.Success
	}

	dep := i.resolver.GetDependency(name)
	if dep == nil {
		return false
	}

	return dep.IsInstalled
}

// GetInstalledVersion returns the installed version of a component.
func (i *DefaultInstaller) GetInstalledVersion(name string) (string, error) {
	if result, ok := i.installed[name]; ok {
		return result.Version, nil
	}

	// Try to detect version
	result, err := i.Verify(context.Background(), name)
	if err != nil {
		return "", err
	}

	return result.Version, nil
}

// reportProgress sends a progress update.
func (i *DefaultInstaller) reportProgress(progress *InstallProgress) {
	progress.UpdatedAt = time.Now()
	select {
	case i.progressChan <- progress:
	default:
		// Channel full, skip update
	}
}

// ensureDirectories creates necessary installation directories.
func (i *DefaultInstaller) ensureDirectories() error {
	dirs := []string{
		i.context.BinDir,
		i.context.FontDir,
		i.context.CacheDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// installBinary installs a binary component.
func (i *DefaultInstaller) installBinary(ctx context.Context, dep *Dependency, opts *Options, result *InstallResult) error {
	i.reportProgress(&InstallProgress{
		Component: dep.Name,
		Stage:     StageDownloading,
		Percent:   10,
		Message:   fmt.Sprintf("Downloading %s...", dep.DisplayName),
	})

	// Download the binary
	downloadResult, err := i.download(ctx, dep, opts)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	i.reportProgress(&InstallProgress{
		Component: dep.Name,
		Stage:     StageVerifying,
		Percent:   50,
		Message:   "Verifying checksum...",
	})

	// Verify checksum if provided
	if dep.Checksum != "" && !opts.SkipChecksum {
		if downloadResult.Checksum != dep.Checksum {
			os.Remove(downloadResult.LocalPath)
			return fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, dep.Checksum, downloadResult.Checksum)
		}
	}

	i.reportProgress(&InstallProgress{
		Component: dep.Name,
		Stage:     StageInstalling,
		Percent:   70,
		Message:   "Installing binary...",
	})

	// Move to final location
	targetPath := filepath.Join(i.context.BinDir, dep.Name)
	if i.context.DryRun {
		result.InstalledPath = targetPath
		return nil
	}

	// Create bin directory if needed
	if err := os.MkdirAll(i.context.BinDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Copy file
	if err := copyFile(downloadResult.LocalPath, targetPath); err != nil {
		return fmt.Errorf("failed to install binary: %w", err)
	}

	// Make executable
	if err := os.Chmod(targetPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	result.InstalledPath = targetPath
	return nil
}

// installFont installs a font component.
func (i *DefaultInstaller) installFont(ctx context.Context, dep *Dependency, opts *Options, result *InstallResult) error {
	i.reportProgress(&InstallProgress{
		Component: dep.Name,
		Stage:     StageDownloading,
		Percent:   10,
		Message:   fmt.Sprintf("Downloading %s...", dep.DisplayName),
	})

	// Download the font
	downloadResult, err := i.download(ctx, dep, opts)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	i.reportProgress(&InstallProgress{
		Component: dep.Name,
		Stage:     StageInstalling,
		Percent:   70,
		Message:   "Installing font...",
	})

	if i.context.DryRun {
		result.InstalledPath = i.context.FontDir
		return nil
	}

	// Ensure font directory exists
	if err := os.MkdirAll(i.context.FontDir, 0755); err != nil {
		return fmt.Errorf("failed to create font directory: %w", err)
	}

	// Extract if zip (fonts often come as zip)
	if strings.HasSuffix(downloadResult.LocalPath, ".zip") {
		return i.installFontFromZip(downloadResult.LocalPath, dep, result)
	}

	// Copy font file directly
	targetPath := filepath.Join(i.context.FontDir, filepath.Base(downloadResult.LocalPath))
	if err := copyFile(downloadResult.LocalPath, targetPath); err != nil {
		return fmt.Errorf("failed to install font: %w", err)
	}

	result.InstalledPath = targetPath

	// Refresh font cache on Linux
	if i.context.OS != "darwin" {
		if err := i.refreshFontCache(); err != nil {
			result.Warnings = append(result.Warnings, "Failed to refresh font cache")
		}
	}

	return nil
}

// installPackage installs a package using the system package manager.
func (i *DefaultInstaller) installPackage(ctx context.Context, dep *Dependency, opts *Options, result *InstallResult) error {
	i.reportProgress(&InstallProgress{
		Component: dep.Name,
		Stage:     StageInstalling,
		Percent:   10,
		Message:   fmt.Sprintf("Installing %s via package manager...", dep.DisplayName),
	})

	if i.context.DryRun {
		return nil
	}

	// Get package name for the current package manager
	packageName := i.getPackageName(dep)
	if packageName == "" {
		return fmt.Errorf("no package mapping for %s on %s", dep.Name, i.context.PackageMgr)
	}

	// Run package manager install command
	var cmd *exec.Cmd
	switch i.context.PackageMgr {
	case "brew":
		cmd = exec.CommandContext(ctx, "brew", "install", packageName)
	case "apt", "apt-get":
		cmd = exec.CommandContext(ctx, "sudo", "apt-get", "install", "-y", packageName)
	case "dnf", "yum":
		cmd = exec.CommandContext(ctx, "sudo", i.context.PackageMgr, "install", "-y", packageName)
	case "pacman":
		cmd = exec.CommandContext(ctx, "sudo", "pacman", "-S", "--noconfirm", packageName)
	case "apk":
		cmd = exec.CommandContext(ctx, "apk", "add", packageName)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedPlatform, i.context.PackageMgr)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("package installation failed: %w\n%s", err, string(output))
	}

	// Find installed path
	path, err := exec.LookPath(dep.Name)
	if err == nil {
		result.InstalledPath = path
	}

	return nil
}

// getPackageName gets the package name for the current package manager.
func (i *DefaultInstaller) getPackageName(dep *Dependency) string {
	// Common package name mappings
	packageMap := map[string]map[string]string{
		"zoxide": {
			"brew":   "zoxide",
			"apt":    "zoxide",
			"pacman": "zoxide",
			"dnf":    "zoxide",
		},
		"fzf": {
			"brew":   "fzf",
			"apt":    "fzf",
			"pacman": "fzf",
			"dnf":    "fzf",
		},
		"bat": {
			"brew":   "bat",
			"apt":    "bat",
			"pacman": "bat",
			"dnf":    "bat",
		},
		"eza": {
			"brew":   "eza",
			"apt":    "eza",
			"pacman": "eza",
			"dnf":    "eza",
		},
	}

	if mappings, ok := packageMap[dep.Name]; ok {
		if name, ok := mappings[i.context.PackageMgr]; ok {
			return name
		}
	}

	return dep.Name // Default to same name
}

// download downloads a file from the dependency source.
func (i *DefaultInstaller) download(ctx context.Context, dep *Dependency, opts *Options) (*DownloadResult, error) {
	// Check cache first
	if opts.UseCache {
		if cachedPath, ok := i.downloadCache[dep.Name]; ok {
			if _, err := os.Stat(cachedPath); err == nil {
				return &DownloadResult{
					URL:       dep.Source,
					LocalPath: cachedPath,
					Cached:    true,
				}, nil
			}
		}
	}

	// Create cache directory
	cacheDir := filepath.Join(i.context.CacheDir, "downloads")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Use the download implementation from download.go
	return i.downloadFile(ctx, dep.Source, cacheDir, opts)
}

// runPostInstallCommands runs post-installation commands.
func (i *DefaultInstaller) runPostInstallCommands(dep *Dependency, result *InstallResult) error {
	for _, cmdStr := range dep.PostInstallCommands {
		// Handle command substitution
		cmdStr = strings.ReplaceAll(cmdStr, "{install_path}", result.InstalledPath)
		cmdStr = strings.ReplaceAll(cmdStr, "{home}", i.context.HomeDir)

		cmdParts := strings.Fields(cmdStr)
		if len(cmdParts) == 0 {
			continue
		}

		cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("post-install command failed: %w\n%s", err, string(output))
		}
	}

	return nil
}

// isComponentInstalled determines if a component is installed.
func (i *DefaultInstaller) isComponentInstalled(dep *Dependency, verifyResult *VerificationResult) bool {
	switch dep.Type {
	case ComponentTypeBinary:
		return verifyResult.InPATH && verifyResult.Working
	case ComponentTypeFont:
		// Check if font file exists
		fontPath := filepath.Join(i.context.FontDir, dep.Name)
		if _, err := os.Stat(fontPath); err == nil {
			return true
		}
		return false
	case ComponentTypePackage:
		// Check if package manager reports it installed
		_, err := exec.LookPath(dep.Name)
		return err == nil
	default:
		return false
	}
}

// uninstallBinary removes a binary component.
func (i *DefaultInstaller) uninstallBinary(dep *Dependency) error {
	targetPath := filepath.Join(i.context.BinDir, dep.Name)
	return os.Remove(targetPath)
}

// uninstallFont removes a font component.
func (i *DefaultInstaller) uninstallFont(dep *Dependency) error {
	// Remove all related font files
	files, err := filepath.Glob(filepath.Join(i.context.FontDir, dep.Name+"*"))
	if err != nil {
		return fmt.Errorf("failed to find font files: %w", err)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("failed to remove font file %s: %w", file, err)
		}
	}

	return nil
}

// extractVersion extracts a version string from output.
func extractVersion(output string) string {
	// Common patterns: "version 1.2.3", "v1.2.3", "1.2.3"
	output = strings.TrimSpace(output)

	// Look for "version X.X.X" pattern
	if strings.Contains(strings.ToLower(output), "version") {
		parts := strings.Fields(output)
		for i, part := range parts {
			if strings.ToLower(part) == "version" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}

	// Look for semantic version pattern in fields
	fields := strings.Fields(output)
	for _, field := range fields {
		if strings.Count(field, ".") >= 1 {
			// Check if it looks like a version
			if len(field) >= 3 && (field[0] >= '0' && field[0] <= '9' || (field[0] == 'v' && len(field) >= 4)) {
				return strings.TrimPrefix(field, "v")
			}
		}
	}

	return output
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0755)
}
