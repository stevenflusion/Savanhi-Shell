// Package persistence provides data persistence for Savanhi Shell.
// This file contains type definitions for persistence data structures.
package persistence

import (
	"time"

	"github.com/savanhi/shell/internal/detector"
)

// OriginalBackup contains the original system state before any modifications.
// This is created once on first run and should never be modified.
type OriginalBackup struct {
	// CreatedAt is when the backup was created.
	CreatedAt time.Time `json:"created_at"`

	// Version is the Savanhi Shell version that created this backup.
	Version string `json:"version"`

	// Shell contains the original shell configuration.
	Shell ShellBackup `json:"shell"`

	// Terminal contains the original terminal configuration.
	Terminal TerminalBackup `json:"terminal"`

	// Fonts contains the original font configuration.
	Fonts FontsBackup `json:"fonts"`

	// Tools contains the original tool configurations.
	Tools ToolsBackup `json:"tools"`

	// RCFiles contains the original RC file contents.
	RCFiles map[string]string `json:"rc_files"`

	// DetectorSnapshot contains the full system detection snapshot.
	DetectorSnapshot *detector.DetectorResult `json:"detector_snapshot"`
}

// ShellBackup contains backup data for shell configuration.
type ShellBackup struct {
	// Name is the original shell name.
	Name string `json:"name"`

	// RCContent is the original RC file content.
	RCContent string `json:"rc_content"`

	// ConfigDirContent contains files from the config directory.
	ConfigDirContent map[string]string `json:"config_dir_content,omitempty"`
}

// TerminalBackup contains backup data for terminal configuration.
type TerminalBackup struct {
	// Name is the terminal name.
	Name string `json:"name"`

	// SettingsFile is the terminal settings file content (if applicable).
	SettingsFile string `json:"settings_file,omitempty"`

	// SettingsPath is the path to the settings file.
	SettingsPath string `json:"settings_path,omitempty"`
}

// FontsBackup contains backup data for font configuration.
type FontsBackup struct {
	// NerdFontsInstalled tracks which Nerd Fonts were installed by Savanhi.
	NerdFontsInstalled []string `json:"nerd_fonts_installed,omitempty"`

	// OriginalFontConfig contains any font configuration files.
	OriginalFontConfig map[string]string `json:"original_font_config,omitempty"`
}

// ToolsBackup contains backup data for installed tools.
type ToolsBackup struct {
	// InstalledBySavanhi tracks which tools were installed by Savanhi.
	InstalledBySavanhi []string `json:"installed_by_savanhi,omitempty"`

	// OriginalConfigs contains original tool configurations.
	OriginalConfigs map[string]string `json:"original_configs,omitempty"`
}

// Preferences contains user preferences for Savanhi Shell.
type Preferences struct {
	// Version is the preferences file format version.
	Version int `json:"version"`

	// LastUpdated is when preferences were last modified.
	LastUpdated time.Time `json:"last_updated"`

	// Theme contains theme preferences.
	Theme ThemePreferences `json:"theme"`

	// Shell contains shell preferences.
	Shell ShellPreferences `json:"shell"`

	// Terminal contains terminal preferences.
	Terminal TerminalPreferences `json:"terminal"`

	// Fonts contains font preferences.
	Fonts FontPreferences `json:"fonts"`

	// Tools contains tool preferences.
	Tools ToolPreferences `json:"tools"`

	// Advanced contains advanced preferences.
	Advanced AdvancedPreferences `json:"advanced"`
}

// ThemePreferences contains theme-related preferences.
type ThemePreferences struct {
	// Name is the selected theme name.
	Name string `json:"name"`

	// Path is the path to the theme file.
	Path string `json:"path,omitempty"`

	// CustomSettings contains theme-specific customizations.
	CustomSettings map[string]interface{} `json:"custom_settings,omitempty"`

	// AutoUpdate enables automatic theme updates.
	AutoUpdate bool `json:"auto_update"`

	// Variant is the theme variant (e.g., "dark", "light").
	Variant string `json:"variant,omitempty"`
}

// ShellPreferences contains shell-related preferences.
type ShellPreferences struct {
	// PreferredShell is the user's preferred shell.
	PreferredShell string `json:"preferred_shell"`

	// EnableSyntaxHighlighting enables syntax highlighting.
	EnableSyntaxHighlighting bool `json:"enable_syntax_highlighting"`

	// EnableAutosuggestions enables command autosuggestions.
	EnableAutosuggestions bool `json:"enable_autosuggestions"`

	// EnableHistorySettings enables history optimization.
	EnableHistorySettings bool `json:"enable_history_settings"`

	// Aliases contains user-defined aliases.
	Aliases map[string]string `json:"aliases,omitempty"`
}

// TerminalPreferences contains terminal-related preferences.
type TerminalPreferences struct {
	// FontFamily is the preferred font family.
	FontFamily string `json:"font_family"`

	// FontSize is the preferred font size.
	FontSize int `json:"font_size"`

	// EnableLigatures enables font ligatures.
	EnableLigatures bool `json:"enable_ligatures"`

	// ColorScheme is the color scheme name.
	ColorScheme string `json:"color_scheme,omitempty"`

	// CursorStyle is the cursor style ("block", "bar", "underline").
	CursorStyle string `json:"cursor_style,omitempty"`
}

// FontPreferences contains font-related preferences.
type FontPreferences struct {
	// PrimaryNerdFont is the primary Nerd Font family.
	PrimaryNerdFont string `json:"primary_nerd_font"`

	// FallbackFont is the fallback font family.
	FallbackFont string `json:"fallback_font,omitempty"`

	// EnableNerdFontIcons enables Nerd Font icons.
	EnableNerdFontIcons bool `json:"enable_nerd_font_icons"`
}

// ToolPreferences contains tool-related preferences.
type ToolPreferences struct {
	// EnableZoxide enables zoxide integration.
	EnableZoxide bool `json:"enable_zoxide"`

	// EnableFzf enables fzf integration.
	EnableFzf bool `json:"enable_fzf"`

	// EnableBat enables bat as cat replacement.
	EnableBat bool `json:"enable_bat"`

	// EnableEza enables eza as ls replacement.
	EnableEza bool `json:"enable_eza"`

	// CustomTools contains custom tool configurations.
	CustomTools map[string]ToolConfig `json:"custom_tools,omitempty"`
}

// ToolConfig contains configuration for a custom tool.
type ToolConfig struct {
	// Enabled indicates if the tool is enabled.
	Enabled bool `json:"enabled"`

	// Version is the installed version.
	Version string `json:"version,omitempty"`

	// ConfigPath is the path to the tool's configuration.
	ConfigPath string `json:"config_path,omitempty"`

	// Settings contains tool-specific settings.
	Settings map[string]interface{} `json:"settings,omitempty"`
}

// AdvancedPreferences contains advanced user preferences.
type AdvancedPreferences struct {
	// CreateBackup enables automatic backup creation.
	CreateBackup bool `json:"create_backup"`

	// BackupRetentionDays is how long to keep backups.
	BackupRetentionDays int `json:"backup_retention_days"`

	// EnableTelemetry enables anonymous usage telemetry.
	EnableTelemetry bool `json:"enable_telemetry"`

	// AutoUpdate enables automatic updates.
	AutoUpdate bool `json:"auto_update"`

	// UpdateChannel is the update channel ("stable", "beta").
	UpdateChannel string `json:"update_channel"`

	// Verbose enables verbose logging.
	Verbose bool `json:"verbose"`

	// Experimental contains experimental feature flags.
	Experimental map[string]bool `json:"experimental,omitempty"`
}

// HistoryEntry represents a single history entry.
type HistoryEntry struct {
	// ID is the unique identifier for this entry.
	ID string `json:"id"`

	// Timestamp is when the action was performed.
	Timestamp time.Time `json:"timestamp"`

	// ActionType is the type of action performed.
	ActionType ActionType `json:"action_type"`

	// Description is a human-readable description.
	Description string `json:"description"`

	// Details contains action-specific details.
	Details map[string]interface{} `json:"details,omitempty"`

	// Status is the action status.
	Status ActionStatus `json:"status"`

	// ErrorMessage contains error details if failed.
	ErrorMessage string `json:"error_message,omitempty"`

	// RollbackAvailable indicates if this action can be rolled back.
	RollbackAvailable bool `json:"rollback_available"`

	// BackupID is the backup ID created before this action.
	BackupID string `json:"backup_id,omitempty"`
}

// ActionType represents the type of action performed.
type ActionType string

const (
	// ActionTypeInstall indicates a package/component installation.
	ActionTypeInstall ActionType = "install"
	// ActionTypeUpdate indicates an update operation.
	ActionTypeUpdate ActionType = "update"
	// ActionTypeRemove indicates a removal operation.
	ActionTypeRemove ActionType = "remove"
	// ActionTypeConfigure indicates a configuration change.
	ActionTypeConfigure ActionType = "configure"
	// ActionTypeRollback indicates a rollback operation.
	ActionTypeRollback ActionType = "rollback"
	// ActionTypeBackup indicates a backup operation.
	ActionTypeBackup ActionType = "backup"
	// ActionTypeRestore indicates a restore operation.
	ActionTypeRestore ActionType = "restore"
)

// ActionStatus represents the status of an action.
type ActionStatus string

const (
	// ActionStatusPending indicates action is pending.
	ActionStatusPending ActionStatus = "pending"
	// ActionStatusInProgress indicates action is in progress.
	ActionStatusInProgress ActionStatus = "in_progress"
	// ActionStatusCompleted indicates action completed successfully.
	ActionStatusCompleted ActionStatus = "completed"
	// ActionStatusFailed indicates action failed.
	ActionStatusFailed ActionStatus = "failed"
	// ActionStatusRolledBack indicates action was rolled back.
	ActionStatusRolledBack ActionStatus = "rolled_back"
)

// Backup is a timestamped backup of system state.
type Backup struct {
	// ID is the unique backup identifier.
	ID string `json:"id"`

	// CreatedAt is when the backup was created.
	CreatedAt time.Time `json:"created_at"`

	// Type is the backup type.
	Type BackupType `json:"type"`

	// Description is a human-readable description.
	Description string `json:"description"`

	// Size is the backup size in bytes.
	Size int64 `json:"size"`

	// Files contains the backup file paths.
	Files []BackupFile `json:"files"`

	// Metadata contains backup metadata.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// BackupType represents the type of backup.
type BackupType string

const (
	// BackupTypeOriginal is the first-run original backup.
	BackupTypeOriginal BackupType = "original"
	// BackupTypeAuto is an automatic backup.
	BackupTypeAuto BackupType = "auto"
	// BackupTypeManual is a user-created backup.
	BackupTypeManual BackupType = "manual"
	// BackupTypePreUpdate is a backup before an update.
	BackupTypePreUpdate BackupType = "pre_update"
)

// BackupFile represents a single file in a backup.
type BackupFile struct {
	// OriginalPath is the original file path.
	OriginalPath string `json:"original_path"`

	// BackupPath is the path in the backup.
	BackupPath string `json:"backup_path"`

	// Hash is the SHA256 hash of the file content.
	Hash string `json:"hash,omitempty"`

	// Size is the file size in bytes.
	Size int64 `json:"size"`
}

// PreviewSession represents an active preview session.
type PreviewSession struct {
	// ID is the unique session identifier.
	ID string `json:"id"`

	// CreatedAt is when the session was created.
	CreatedAt time.Time `json:"created_at"`

	// Theme is the theme being previewed.
	Theme string `json:"theme"`

	// RCBackup is the backup of the RC file before preview.
	RCBackup string `json:"rc_backup"`

	// SubshellPID is the PID of the preview subshell.
	SubshellPID int `json:"subshell_pid,omitempty"`

	// Active indicates if the session is still active.
	Active bool `json:"active"`
}
