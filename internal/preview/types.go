// Package preview provides live preview capabilities for Savanhi Shell.
// It spawns isolated subshells to preview themes, fonts, and colors.
package preview

import (
	"context"
	"time"

	"github.com/savanhi/shell/pkg/shell"
)

// PreviewType represents the type of preview.
type PreviewType string

const (
	// PreviewTypeTheme is for theme previews.
	PreviewTypeTheme PreviewType = "theme"
	// PreviewTypeFont is for font previews.
	PreviewTypeFont PreviewType = "font"
	// PreviewTypeColorScheme is for color scheme previews.
	PreviewTypeColorScheme PreviewType = "color_scheme"
	// PreviewTypeFull is for combined previews (theme + font + colors).
	PreviewTypeFull PreviewType = "full"
)

// PreviewStatus represents the status of a preview session.
type PreviewStatus string

const (
	// StatusPending indicates the preview is pending.
	StatusPending PreviewStatus = "pending"
	// StatusRunning indicates the preview is running.
	StatusRunning PreviewStatus = "running"
	// StatusCompleted indicates the preview completed successfully.
	StatusCompleted PreviewStatus = "completed"
	// StatusFailed indicates the preview failed.
	StatusFailed PreviewStatus = "failed"
	// StatusCancelled indicates the preview was cancelled.
	StatusCancelled PreviewStatus = "cancelled"
	// StatusTimeout indicates the preview timed out.
	StatusTimeout PreviewStatus = "timeout"
)

// PreviewConfig contains configuration for a preview session.
type PreviewConfig struct {
	// Type is the preview type (theme, font, color, full).
	Type PreviewType `json:"type"`

	// Shell is the target shell for the preview.
	Shell shell.ShellType `json:"shell"`

	// ThemePath is the path to the Oh My Posh theme file.
	// Required for theme and full previews.
	ThemePath string `json:"theme_path,omitempty"`

	// ThemeName is the name of the theme (for display purposes).
	ThemeName string `json:"theme_name,omitempty"`

	// FontFamily is the font family to preview.
	// Required for font and full previews.
	FontFamily string `json:"font_family,omitempty"`

	// FontSize is the font size in points.
	FontSize int `json:"font_size,omitempty"`

	// ColorScheme is the color scheme name.
	// Required for color scheme and full previews.
	ColorScheme string `json:"color_scheme,omitempty"`

	// Environment variables to inject into the preview subshell.
	Environment map[string]string `json:"environment,omitempty"`

	// Timeout is the maximum duration for the preview.
	// Default is 5 seconds if not specified.
	Timeout time.Duration `json:"timeout,omitempty"`

	// WorkingDir is the working directory for the preview.
	// Defaults to user's home directory if not specified.
	WorkingDir string `json:"working_dir,omitempty"`

	// CaptureOutput indicates whether to capture stdout/stderr.
	// Default is true.
	CaptureOutput bool `json:"capture_output"`

	// PreviewCommand is a custom command to run for the preview.
	// If empty, defaults to echoing the prompt.
	PreviewCommand string `json:"preview_command,omitempty"`
}

// PreviewResult contains the result of a preview session.
type PreviewResult struct {
	// ID is the unique identifier for this preview.
	ID string `json:"id"`

	// Status is the preview status.
	Status PreviewStatus `json:"status"`

	// Output is the captured stdout from the preview.
	Output string `json:"output,omitempty"`

	// ErrorOutput is the captured stderr from the preview.
	ErrorOutput string `json:"error_output,omitempty"`

	// ExitCode is the process exit code.
	ExitCode int `json:"exit_code"`

	// ErrorMessage contains error details if the preview failed.
	ErrorMessage string `json:"error_message,omitempty"`

	// Duration is how long the preview took.
	Duration time.Duration `json:"duration"`

	// StartTime is when the preview started.
	StartTime time.Time `json:"start_time"`

	// EndTime is when the preview ended.
	EndTime time.Time `json:"end_time"`

	// Config is the preview configuration that was used.
	Config PreviewConfig `json:"config"`
}

// PreviewSessionState represents the state of an active preview session.
type PreviewSessionState struct {
	// ID is the unique session identifier.
	ID string `json:"id"`

	// Config is the preview configuration.
	Config PreviewConfig `json:"config"`

	// Status is the current status.
	Status PreviewStatus `json:"status"`

	// PID is the process ID of the subshell.
	PID int `json:"pid,omitempty"`

	// TempDir is the temporary directory created for this session.
	TempDir string `json:"temp_dir,omitempty"`

	// TempRCFile is the temporary RC file path.
	TempRCFile string `json:"temp_rc_file,omitempty"`

	// StartTime is when the session started.
	StartTime time.Time `json:"start_time"`

	// CancelFunc is the context cancel function for cleanup.
	// Not serialized - used internally.
	CancelFunc context.CancelFunc `json:"-"`

	// OutputBuffer is the buffer for captured output.
	// Not serialized - used internally.
	OutputBuffer []byte `json:"-"`
}

// SubshellConfig contains configuration for subshell spawning.
type SubshellConfig struct {
	// ShellType is the shell to spawn (bash or zsh).
	ShellType shell.ShellType `json:"shell_type"`

	// ShellPath is the path to the shell executable.
	// If empty, detected automatically.
	ShellPath string `json:"shell_path,omitempty"`

	// Environment variables to set in the subshell.
	Environment map[string]string `json:"environment,omitempty"`

	// Timeout is the maximum duration for the subshell.
	// Default is 5 seconds.
	Timeout time.Duration `json:"timeout,omitempty"`

	// WorkingDir is the working directory for the subshell.
	// Defaults to home directory if not specified.
	WorkingDir string `json:"working_dir,omitempty"`

	// RCContent is the content for the temporary RC file.
	// If empty, a minimal RC is generated.
	RCContent string `json:"rc_content,omitempty"`

	// Command is the command to execute in the subshell.
	// If empty, an interactive shell is spawned.
	Command string `json:"command,omitempty"`

	// CaptureStdout indicates whether to capture stdout.
	CaptureStdout bool `json:"capture_stdout"`

	// CaptureStderr indicates whether to capture stderr.
	CaptureStderr bool `json:"capture_stderr"`
}

// SubshellResult contains the result of a subshell execution.
type SubshellResult struct {
	// PID is the process ID of the spawned subshell.
	PID int `json:"pid"`

	// Stdout is the captured standard output.
	Stdout string `json:"stdout,omitempty"`

	// Stderr is the captured standard error.
	Stderr string `json:"stderr,omitempty"`

	// ExitCode is the process exit code.
	ExitCode int `json:"exit_code"`

	// Duration is how long the subshell ran.
	Duration time.Duration `json:"duration"`

	// TempFilePaths are paths to temporary files created.
	TempFilePaths []string `json:"temp_file_paths,omitempty"`
}

// EnvironmentInjection contains environment variables for preview.
type EnvironmentInjection struct {
	// OhMyPosh theme configuration.
	POSHTheme string `json:"posh_theme,omitempty"`

	// Font family for terminal rendering.
	FontFamily string `json:"font_family,omitempty"`

	// Font size in points.
	FontSize string `json:"font_size,omitempty"`

	// Terminal type for color support detection.
	TERM string `json:"term,omitempty"`

	// Color scheme identifier.
	ColorScheme string `json:"color_scheme,omitempty"`

	// Additional environment variables.
	Extra map[string]string `json:"extra,omitempty"`
}

// ThemePreviewConfig contains configuration for theme preview.
type ThemePreviewConfig struct {
	// ThemeName is the display name of the theme.
	ThemeName string `json:"theme_name"`

	// ThemePath is the path to the theme JSON file.
	ThemePath string `json:"theme_path"`

	// ThemeContent is the raw theme content (for bundled themes).
	ThemeContent string `json:"theme_content,omitempty"`

	// Shell is the target shell type.
	Shell shell.ShellType `json:"shell"`

	// OhMyPoshPath is the path to oh-my-posh binary.
	// If empty, detected from PATH.
	OhMyPoshPath string `json:"oh_my_posh_path,omitempty"`

	// Timeout is the preview timeout.
	Timeout time.Duration `json:"timeout,omitempty"`
}

// FontPreviewConfig contains configuration for font preview.
type FontPreviewConfig struct {
	// FontFamily is the font family name.
	FontFamily string `json:"font_family"`

	// FontSize is the font size in points.
	FontSize int `json:"font_size,omitempty"`

	// SampleText is custom text to display for preview.
	SampleText string `json:"sample_text,omitempty"`

	// ShowNerdFontIcons indicates whether to show Nerd Font icons.
	ShowNerdFontIcons bool `json:"show_nerd_font_icons"`

	// Shell is the target shell type.
	Shell shell.ShellType `json:"shell"`

	// Timeout is the preview timeout.
	Timeout time.Duration `json:"timeout,omitempty"`
}

// ColorSchemePreviewConfig contains configuration for color scheme preview.
type ColorSchemePreviewConfig struct {
	// ColorSchemeName is the name of the color scheme.
	ColorSchemeName string `json:"color_scheme_name"`

	// ColorValues contains the color values for the scheme.
	ColorValues map[string]string `json:"color_values,omitempty"`

	// ShowPalette indicates whether to show the full color palette.
	ShowPalette bool `json:"show_palette"`

	// Shell is the target shell type.
	Shell shell.ShellType `json:"shell"`

	// Timeout is the preview timeout.
	Timeout time.Duration `json:"timeout,omitempty"`
}

// CleanupResult contains the result of cleanup operations.
type CleanupResult struct {
	// RemovedFiles are files that were removed.
	RemovedFiles []string `json:"removed_files,omitempty"`

	// KilledProcesses are PIDs that were killed.
	KilledProcesses []int `json:"killed_processes,omitempty"`

	// Errors are any errors encountered during cleanup.
	Errors []string `json:"errors,omitempty"`

	// Success indicates if cleanup completed without errors.
	Success bool `json:"success"`
}

// DefaultTimeout is the default timeout for preview operations.
const DefaultTimeout = 5 * time.Second

// DefaultPreviewCommand is the default command for generating prompt output.
const DefaultPreviewCommand = "echo $PROMPT"
