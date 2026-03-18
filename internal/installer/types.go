// Package installer provides dependency installation and management for Savanhi Shell.
// This file contains type definitions for the installer module.
package installer

import (
	"context"
	"time"
)

// ComponentType represents the type of installable component.
type ComponentType string

const (
	// ComponentTypeBinary is a downloadable binary.
	ComponentTypeBinary ComponentType = "binary"
	// ComponentTypeFont is a Nerd Font.
	ComponentTypeFont ComponentType = "font"
	// ComponentTypePackage is a system package.
	ComponentTypePackage ComponentType = "package"
	// ComponentTypeConfig is a configuration file.
	ComponentTypeConfig ComponentType = "config"
)

// Platform represents an operating system platform.
type Platform string

const (
	// PlatformMacOS represents macOS.
	PlatformMacOS Platform = "darwin"
	// PlatformLinux represents Linux.
	PlatformLinux Platform = "linux"
	// PlatformWindows represents Windows.
	PlatformWindows Platform = "windows"
	// PlatformWSL represents Windows Subsystem for Linux.
	PlatformWSL Platform = "wsl"
	// PlatformTermux represents Android Termux.
	PlatformTermux Platform = "termux"
)

// Dependency represents a single installable dependency.
type Dependency struct {
	// Name is the dependency name (e.g., "oh-my-posh", "zoxide").
	Name string `json:"name"`

	// DisplayName is a human-readable name.
	DisplayName string `json:"display_name"`

	// Description is a brief description of what this dependency provides.
	Description string `json:"description"`

	// Version is the required or installed version.
	Version string `json:"version"`

	// Type is the component type.
	Type ComponentType `json:"type"`

	// Source is the download URL or package source.
	Source string `json:"source"`

	// Checksum is the expected SHA256 checksum.
	Checksum string `json:"checksum,omitempty"`

	// ChecksumURL is the URL to fetch the checksum from.
	ChecksumURL string `json:"checksum_url,omitempty"`

	// Dependencies are other dependencies that must be installed first.
	Dependencies []string `json:"dependencies,omitempty"`

	// Platforms are the supported platforms (empty = all platforms).
	Platforms []Platform `json:"platforms,omitempty"`

	// InstallPath is where the component will be installed.
	InstallPath string `json:"install_path,omitempty"`

	// IsInstalled indicates if the dependency is already installed.
	IsInstalled bool `json:"is_installed,omitempty"`

	// InstalledVersion is the currently installed version.
	InstalledVersion string `json:"installed_version,omitempty"`

	// Optional indicates if this dependency is optional.
	Optional bool `json:"optional"`

	// PostInstallCommands are commands to run after installation.
	PostInstallCommands []string `json:"post_install_commands,omitempty"`

	// VerifyCommand is the command to verify installation.
	VerifyCommand string `json:"verify_command,omitempty"`

	// UninstallCommand is the command to uninstall.
	UninstallCommand string `json:"uninstall_command,omitempty"`
}

// InstallProgress represents progress information for an installation.
type InstallProgress struct {
	// Component is the name of the component being installed.
	Component string `json:"component"`

	// Stage is the current installation stage.
	Stage InstallStage `json:"stage"`

	// Percent is the completion percentage (0-100).
	Percent float64 `json:"percent"`

	// BytesDownloaded is the number of bytes downloaded.
	BytesDownloaded int64 `json:"bytes_downloaded"`

	// TotalBytes is the total number of bytes to download.
	TotalBytes int64 `json:"total_bytes"`

	// Message is a human-readable progress message.
	Message string `json:"message"`

	// Error contains any error message if failed.
	Error string `json:"error,omitempty"`

	// StartTime is when this progress started.
	StartTime time.Time `json:"start_time"`

	// UpdatedAt is when this progress was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// InstallStage represents the current stage of installation.
type InstallStage string

const (
	// StageIdle indicates no installation in progress.
	StageIdle InstallStage = "idle"
	// StageResolving indicates dependency resolution in progress.
	StageResolving InstallStage = "resolving"
	// StageDownloading indicates download in progress.
	StageDownloading InstallStage = "downloading"
	// StageVerifying indicates checksum verification in progress.
	StageVerifying InstallStage = "verifying"
	// StageInstalling indicates installation in progress.
	StageInstalling InstallStage = "installing"
	// StageConfiguring indicates configuration in progress.
	StageConfiguring InstallStage = "configuring"
	// StageVerifyingInstall indicates installation verification in progress.
	StageVerifyingInstall InstallStage = "verifying_install"
	// StageCompleted indicates installation completed successfully.
	StageCompleted InstallStage = "completed"
	// StageFailed indicates installation failed.
	StageFailed InstallStage = "failed"
	// StageRolledBack indicates installation was rolled back.
	StageRolledBack InstallStage = "rolled_back"
)

// InstallResult represents the result of an installation.
type InstallResult struct {
	// Component is the name of the installed component.
	Component string `json:"component"`

	// Success indicates if installation was successful.
	Success bool `json:"success"`

	// Version is the installed version.
	Version string `json:"version"`

	// InstalledPath is the path where the component was installed.
	InstalledPath string `json:"installed_path"`

	// Duration is how long installation took.
	Duration time.Duration `json:"duration"`

	// Error contains error details if failed.
	Error string `json:"error,omitempty"`

	// Warnings are non-fatal issues encountered.
	Warnings []string `json:"warnings,omitempty"`

	// RequiresRestart indicates if a shell restart is needed.
	RequiresRestart bool `json:"requires_restart"`
}

// DownloadResult represents the result of a download operation.
type DownloadResult struct {
	// URL is the source URL.
	URL string `json:"url"`

	// LocalPath is the local file path.
	LocalPath string `json:"local_path"`

	// Size is the file size in bytes.
	Size int64 `json:"size"`

	// Checksum is the SHA256 checksum of the downloaded file.
	Checksum string `json:"checksum"`

	// Verified indicates if checksum verification passed.
	Verified bool `json:"verified"`

	// Cached indicates if the file was served from cache.
	Cached bool `json:"cached"`

	// Duration is how long the download took.
	Duration time.Duration `json:"duration"`
}

// VerificationResult represents the result of a verification check.
type VerificationResult struct {
	// Component is the name of the verified component.
	Component string `json:"component"`

	// Installed indicates if the component is installed.
	Installed bool `json:"installed"`

	// Version is the installed version (if any).
	Version string `json:"version,omitempty"`

	// Path is the installation path (if any).
	Path string `json:"path,omitempty"`

	// InPATH indicates if the binary is in PATH.
	InPATH bool `json:"in_path"`

	// Working indicates if the component is functioning correctly.
	Working bool `json:"working"`

	// Issues are any problems detected.
	Issues []string `json:"issues,omitempty"`

	// Checks are individual check results.
	Checks []VerificationCheck `json:"checks,omitempty"`
}

// VerificationCheck represents a single verification check.
type VerificationCheck struct {
	// Name is the check name.
	Name string `json:"name"`

	// Passed indicates if the check passed.
	Passed bool `json:"passed"`

	// Message describes the check result.
	Message string `json:"message,omitempty"`
}

// RollbackResult represents the result of a rollback operation.
type RollbackResult struct {
	// Success indicates if rollback was successful.
	Success bool `json:"success"`

	// Components are the components that were rolled back.
	Components []string `json:"components"`

	// Files are the files that were restored.
	Files []string `json:"files"`

	// Error contains error details if rollback failed.
	Error string `json:"error,omitempty"`

	// Warnings are non-fatal issues encountered during rollback.
	Warnings []string `json:"warnings,omitempty"`

	// Duration is how long the rollback took.
	Duration time.Duration `json:"duration"`
}

// DependencyStatus represents the installation status of dependencies.
type DependencyStatus struct {
	// Name is the dependency name.
	Name string `json:"name"`

	// Installed indicates if it's installed.
	Installed bool `json:"installed"`

	// Version is the installed version.
	Version string `json:"version,omitempty"`

	// Required is the required version.
	Required string `json:"required,omitempty"`

	// Satisfied indicates if the requirement is satisfied.
	Satisfied bool `json:"satisfied"`

	// Missing indicates if the dependency is missing.
	Missing bool `json:"missing"`

	// Outdated indicates if an outdated version is installed.
	Outdated bool `json:"outdated"`

	// InstallPath is where it's installed (if installed).
	InstallPath string `json:"install_path,omitempty"`
}

// InstallPlan represents a planned installation sequence.
type InstallPlan struct {
	// Components are the components to install in order.
	Components []*Dependency `json:"components"`

	// TotalSize is the estimated total download size.
	TotalSize int64 `json:"total_size"`

	// EstimatedDuration is the estimated installation time.
	EstimatedDuration time.Duration `json:"estimated_duration"`

	// RequiresDownload indicates if any downloads are needed.
	RequiresDownload bool `json:"requires_download"`

	// RequiresRestart indicates if a shell restart is needed.
	RequiresRestart bool `json:"requires_restart"`

	// Conflicts are any detected conflicts.
	Conflicts []string `json:"conflicts,omitempty"`

	// Warnings are any warnings about the installation.
	Warnings []string `json:"warnings,omitempty"`
}

// InstallContext provides context for installation operations.
type InstallContext struct {
	// Context is the Go context for cancellation.
	Context context.Context `json:"-"`

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

	// Arch is the detected architecture.
	Arch string `json:"arch"`

	// PackageMgr is the detected package manager.
	PackageMgr string `json:"package_manager"`

	// Shell is the detected shell.
	Shell string `json:"shell"`

	// DryRun indicates if this is a dry run (no actual changes).
	DryRun bool `json:"dry_run"`

	// Force indicates if existing installations should be overwritten.
	Force bool `json:"force"`

	// Verbose enables verbose output.
	Verbose bool `json:"verbose"`

	// NoProgress disables progress reporting.
	NoProgress bool `json:"no_progress"`
}

// Options contains optional installation settings.
type Options struct {
	// SkipChecksum skips checksum verification.
	SkipChecksum bool `json:"skip_checksum"`

	// SkipVerification skips post-install verification.
	SkipVerification bool `json:"skip_verification"`

	// UseCache uses cached downloads if available.
	UseCache bool `json:"use_cache"`

	// MaxRetries is the maximum number of download retries.
	MaxRetries int `json:"max_retries"`

	// Timeout is the download timeout.
	Timeout time.Duration `json:"timeout"`

	// ProgressCallback is called with progress updates.
	ProgressCallback func(progress *InstallProgress) `json:"-"`

	// ConcurrentDownloads enables parallel downloads.
	ConcurrentDownloads bool `json:"concurrent_downloads"`
}

// DefaultOptions returns default installation options.
func DefaultOptions() *Options {
	return &Options{
		SkipChecksum:        false,
		SkipVerification:    false,
		UseCache:            true,
		MaxRetries:          3,
		Timeout:             5 * time.Minute,
		ConcurrentDownloads: true,
	}
}
