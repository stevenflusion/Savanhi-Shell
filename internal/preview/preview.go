// Package preview provides live preview capabilities for Savanhi Shell.
// This file defines the Previewer interface for spawning isolated subshells.
package preview

import (
	"context"
	"time"
)

// Previewer is the interface for spawning isolated preview subshells.
// Implementations provide safe, isolated environments for previewing
// themes, fonts, and color schemes without modifying the user's actual
// shell configuration.
type Previewer interface {
	// Preview spawns an isolated subshell with the given configuration.
	// The preview runs in an isolated environment that does not affect
	// the user's actual shell configuration.
	//
	// The context is used for timeout enforcement and cancellation.
	// When the context is cancelled, the subshell process is terminated.
	//
	// Returns a PreviewResult containing the output and status.
	// Returns an error if the preview cannot be started or fails critically.
	Preview(ctx context.Context, config *PreviewConfig) (*PreviewResult, error)

	// Cleanup releases all resources associated with a preview session.
	// This includes:
	//   - Killing any remaining subshell processes
	//   - Removing temporary files (RC files, theme files, etc.)
	//   - Cleaning up temporary directories
	//
	// Cleanup should be called after every preview, even if the preview
	// completed successfully or failed. It is safe to call Cleanup multiple
	// times.
	Cleanup(sessionID string) (*CleanupResult, error)

	// CleanupAll releases resources for all preview sessions.
	// This is useful for cleanup on application exit.
	CleanupAll() (*CleanupResult, error)

	// GetActiveSessions returns all currently active preview sessions.
	GetActiveSessions() ([]*PreviewSessionState, error)

	// GetSession returns a specific preview session by ID.
	GetSession(id string) (*PreviewSessionState, error)

	// Cancel cancels a running preview session.
	Cancel(sessionID string) error
}

// ThemePreview provides theme preview functionality.
type ThemePreview interface {
	// PreviewTheme generates a preview for a specific theme.
	// It downloads or loads the theme configuration, creates a temporary
	// RC file with oh-my-posh initialization, and spawns a preview subshell.
	PreviewTheme(ctx context.Context, config *ThemePreviewConfig) (*PreviewResult, error)

	// ListAvailableThemes returns a list of available themes.
	// This includes bundled themes and any user-installed themes.
	ListAvailableThemes() ([]string, error)

	// GetThemePath returns the path to a theme file.
	// For bundled themes, this returns the path within the embedded themes.
	// For custom themes, this returns the user-specified path.
	GetThemePath(themeName string) (string, error)
}

// FontPreview provides font preview functionality.
type FontPreview interface {
	// PreviewFont generates a preview for a specific font.
	// It displays sample characters and Nerd Font icons if available.
	PreviewFont(ctx context.Context, config *FontPreviewConfig) (*PreviewResult, error)

	// CheckFontAvailability checks if a font is installed on the system.
	CheckFontAvailability(fontFamily string) (bool, error)

	// GetInstalledFonts returns a list of installed fonts.
	// Optionally filters for Nerd Fonts only.
	GetInstalledFonts(nerdFontsOnly bool) ([]string, error)
}

// ColorSchemePreview provides color scheme preview functionality.
type ColorSchemePreview interface {
	// PreviewColorScheme generates a preview for a color scheme.
	// It shows the color palette and sample output with the scheme applied.
	PreviewColorScheme(ctx context.Context, config *ColorSchemePreviewConfig) (*PreviewResult, error)

	// GetTerminalColorCapabilities returns the terminal's color support.
	// Returns true if the terminal supports true color (24-bit).
	GetTerminalColorCapabilities() (supportsTrueColor bool, supports256Color bool, err error)

	// ListAvailableColorSchemes returns a list of built-in color schemes.
	ListAvailableColorSchemes() ([]string, error)
}

// SubshellSpawner provides subshell spawning functionality.
// This is the core capability used by all preview types.
type SubshellSpawner interface {
	// Spawn creates a new subshell process with the given configuration.
	// The subshell runs in isolation with environment variables injected
	// and a temporary RC file.
	//
	// Returns a SubshellResult with process info and captured output.
	// The caller is responsible for calling Kill() to terminate the process.
	Spawn(ctx context.Context, config *SubshellConfig) (*SubshellResult, error)

	// Kill terminates a subshell process by PID.
	// It sends SIGTERM first, then SIGKILL after a grace period.
	Kill(pid int) error

	// KillAll terminates all subshell processes spawned by this spawner.
	KillAll() error
}

// EnvironmentInjector provides environment variable injection for previews.
type EnvironmentInjector interface {
	// InjectEnvironment creates environment variables for a preview.
	// It combines system-detected values with preview-specific values.
	InjectEnvironment(config *PreviewConfig) (map[string]string, error)

	// InjectThemeEnv injects theme-specific environment variables.
	InjectThemeEnv(themePath string) map[string]string

	// InjectFontEnv injects font-specific environment variables.
	InjectFontEnv(fontFamily string, fontSize int) map[string]string

	// InjectColorSchemeEnv injects color scheme environment variables.
	InjectColorSchemeEnv(schemeName string) map[string]string
}

// SessionManager manages preview sessions across the application.
type SessionManager interface {
	// CreateSession creates a new preview session.
	// Returns an error if a session is already active.
	CreateSession(config *PreviewConfig) (*PreviewSessionState, error)

	// EndSession ends an active preview session.
	// Cleans up all associated resources.
	EndSession(sessionID string) error

	// GetActiveSession returns the currently active session, if any.
	GetActiveSession() (*PreviewSessionState, error)

	// HasActiveSession checks if there's an active session.
	HasActiveSession() bool

	// UpdateSessionStatus updates the status of a session.
	UpdateSessionStatus(sessionID string, status PreviewStatus) error
}

// PreviewSafety provides safety mechanisms for preview operations.
type PreviewSafety interface {
	// EnforceTimeout ensures the preview does not exceed the timeout.
	// It returns a context that will be cancelled after the timeout.
	EnforceTimeout(parentCtx context.Context, timeout time.Duration) (context.Context, context.CancelFunc)

	// RecoverPanic recovers from panics in preview goroutines.
	// Should be called with defer in all preview goroutines.
	RecoverPanic() func()

	// ValidateConfig validates a preview configuration.
	// Returns an error if the config is invalid or unsafe.
	ValidateConfig(config *PreviewConfig) error
}
