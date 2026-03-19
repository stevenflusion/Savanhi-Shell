// Package installer provides dependency installation and management for Savanhi Shell.
// This file contains type definitions for zsh plugin management.
package installer

// InstallMethod represents how a zsh plugin was or will be installed.
type InstallMethod int

const (
	// MethodNone indicates the plugin is not installed.
	MethodNone InstallMethod = iota
	// MethodOhMyZsh indicates installation via Oh My Zsh plugins array.
	MethodOhMyZsh
	// MethodHomebrew indicates installation via Homebrew package.
	MethodHomebrew
	// MethodGitClone indicates installation via manual git clone.
	MethodGitClone
)

// String returns a human-readable representation of the install method.
func (m InstallMethod) String() string {
	switch m {
	case MethodNone:
		return "none"
	case MethodOhMyZsh:
		return "Oh My Zsh"
	case MethodHomebrew:
		return "Homebrew"
	case MethodGitClone:
		return "Git Clone"
	default:
		return "unknown"
	}
}

// Plugin represents a zsh plugin definition.
type Plugin struct {
	// Name is the plugin identifier (e.g., "zsh-autosuggestions").
	Name string `json:"name"`

	// DisplayName is a human-readable name (e.g., "Zsh Autosuggestions").
	DisplayName string `json:"display_name"`

	// Description is a brief description of what the plugin provides.
	Description string `json:"description"`

	// Repository is the Git repository URL for cloning.
	Repository string `json:"repository"`

	// BrewPackage is the Homebrew package name (if available).
	BrewPackage string `json:"brew_package,omitempty"`

	// SourceFile is the main .zsh file to source (e.g., "zsh-autosuggestions.zsh").
	SourceFile string `json:"source_file"`

	// OhMyZshName is the plugin name for OMZ plugins array (may differ from Name).
	OhMyZshName string `json:"ohmyzsh_name,omitempty"`

	// MinZshVersion is the minimum required zsh version (e.g., "4.3.11").
	MinZshVersion string `json:"min_zsh_version,omitempty"`

	// MustBeLast indicates this plugin must be sourced last (e.g., zsh-syntax-highlighting).
	MustBeLast bool `json:"must_be_last"`
}

// PluginStatus represents the detected status of a plugin.
type PluginStatus struct {
	// Plugin is the plugin definition.
	Plugin Plugin `json:"plugin"`

	// Installed indicates whether the plugin is currently installed.
	Installed bool `json:"installed"`

	// Method is how the plugin was installed (if installed).
	Method InstallMethod `json:"method"`

	// InstallPath is where the plugin is installed (if installed).
	InstallPath string `json:"install_path,omitempty"`

	// Version is the installed version (if detectable).
	Version string `json:"version,omitempty"`

	// Conflicts lists any conflicting plugins or plugin managers detected.
	Conflicts []string `json:"conflicts,omitempty"`
}

// PluginInstallResult represents the result of a plugin installation attempt.
type PluginInstallResult struct {
	// Plugin is the plugin that was installed.
	Plugin Plugin `json:"plugin"`

	// Success indicates whether installation succeeded.
	Success bool `json:"success"`

	// Method is the installation method used.
	Method InstallMethod `json:"method"`

	// InstallPath is where the plugin was installed.
	InstallPath string `json:"install_path,omitempty"`

	// RCModified indicates whether .zshrc was modified.
	RCModified bool `json:"rc_modified"`

	// Error contains error details if installation failed.
	Error error `json:"error,omitempty"`

	// Warnings are non-fatal issues encountered.
	Warnings []string `json:"warnings,omitempty"`
}

// PluginInstallerConfig holds configuration for plugin installation.
type PluginInstallerConfig struct {
	// PreferOhMyZsh indicates preference for Oh My Zsh installation when available.
	PreferOhMyZsh bool `json:"prefer_ohmyzsh"`

	// PreferHomebrew indicates preference for Homebrew installation over git clone.
	PreferHomebrew bool `json:"prefer_homebrew"`

	// Force allows reinstallation of already installed plugins.
	Force bool `json:"force"`

	// DryRun simulates installation without making changes.
	DryRun bool `json:"dry_run"`

	// CustomPluginDir is a custom directory for git clone installations.
	CustomPluginDir string `json:"custom_plugin_dir,omitempty"`
}

// DefaultPluginInstallerConfig returns sensible defaults for plugin installation.
func DefaultPluginInstallerConfig() *PluginInstallerConfig {
	return &PluginInstallerConfig{
		PreferOhMyZsh:  true,
		PreferHomebrew: true,
		Force:          false,
		DryRun:         false,
	}
}

// GetSupportedPlugins returns the list of supported zsh plugins.
// These are the predefined plugins that Savanhi can install/manage.
func GetSupportedPlugins() []Plugin {
	return []Plugin{
		{
			Name:          "zsh-autosuggestions",
			DisplayName:   "Zsh Autosuggestions",
			Description:   "Fish-like fast/unobtrusive autosuggestions for zsh",
			Repository:    "https://github.com/zsh-users/zsh-autosuggestions",
			BrewPackage:   "zsh-autosuggestions",
			SourceFile:    "zsh-autosuggestions.zsh",
			OhMyZshName:   "zsh-autosuggestions",
			MinZshVersion: "4.3.11",
			MustBeLast:    false,
		},
		{
			Name:          "zsh-syntax-highlighting",
			DisplayName:   "Zsh Syntax Highlighting",
			Description:   "Fish-like syntax highlighting for zsh. Must be loaded last.",
			Repository:    "https://github.com/zsh-users/zsh-syntax-highlighting",
			BrewPackage:   "zsh-syntax-highlighting",
			SourceFile:    "zsh-syntax-highlighting.zsh",
			OhMyZshName:   "zsh-syntax-highlighting",
			MinZshVersion: "4.3.11",
			MustBeLast:    true,
		},
	}
}

// GetPluginByName returns a plugin by its name, or nil if not found.
func GetPluginByName(name string) *Plugin {
	for _, plugin := range GetSupportedPlugins() {
		if plugin.Name == name {
			return &plugin
		}
	}
	return nil
}
