// Package preview provides live preview capabilities for Savanhi Shell.
// This file implements color scheme preview functionality.
package preview

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Color scheme preview constants.
const (
	// DefaultColorTimeout is the default timeout for color previews.
	DefaultColorTimeout = 3 * time.Second

	// ColorRange16 is the 16-color palette.
	ColorRange16 = "16"

	// ColorRange256 is the 256-color palette.
	ColorRange256 = "256"

	// ColorRangeTrueColor is the true color (24-bit) palette.
	ColorRangeTrueColor = "truecolor"
)

// Common errors for color scheme preview operations.
var (
	// ErrColorSchemeNotFound indicates the color scheme was not found.
	ErrColorSchemeNotFound = errors.New("color scheme not found")

	// ErrColorPreviewFailed indicates the color preview failed.
	ErrColorPreviewFailed = errors.New("color preview failed")

	// ErrTerminalNotSupported indicates the terminal doesn't support colors.
	ErrTerminalNotSupported = errors.New("terminal does not support required colors")
)

// ColorScheme represents a color scheme definition.
type ColorScheme struct {
	// Name is the scheme name.
	Name string `json:"name"`

	// DisplayName is the human-readable name.
	DisplayName string `json:"display_name"`

	// Description is a brief description.
	Description string `json:"description,omitempty"`

	// Colors contains the color definitions.
	Colors ColorDefinitions `json:"colors"`

	// TerminalColors are ANSI color values.
	TerminalColors map[string]string `json:"terminal_colors,omitempty"`
}

// ColorDefinitions contains all color definitions.
type ColorDefinitions struct {
	// Primary colors.
	Foreground string `json:"foreground"`
	Background string `json:"background"`

	// Normal colors (8 standard colors).
	Normal ColorPalette `json:"normal"`

	// Bright colors (8 bright colors).
	Bright ColorPalette `json:"bright"`

	// Extended colors for 256/true color terminals.
	Extended ExtendedColors `json:"extended,omitempty"`
}

// ColorPalette contains a standard 8-color palette.
type ColorPalette struct {
	Black   string `json:"black"`
	Red     string `json:"red"`
	Green   string `json:"green"`
	Yellow  string `json:"yellow"`
	Blue    string `json:"blue"`
	Magenta string `json:"magenta"`
	Cyan    string `json:"cyan"`
	White   string `json:"white"`
}

// ExtendedColors contains extended color definitions.
type ExtendedColors struct {
	// Additional colors for true color terminals.
	Additional map[string]string `json:"additional,omitempty"`
}

// ColorSupport represents terminal color capabilities.
type ColorSupport struct {
	// SupportsBasic indicates basic 16-color support.
	SupportsBasic bool

	// Supports256 indicates 256-color support.
	Supports256 bool

	// SupportsTrueColor indicates 24-bit true color support.
	SupportsTrueColor bool
}

// DefaultColorSchemeHandler implements ColorSchemePreview interface.
type DefaultColorSchemeHandler struct {
	// subsheller is the subshell spawner.
	subsheller SubshellSpawner

	// injector is the environment injector.
	injector EnvironmentInjector

	// cachedSchemes is the cache of loaded color schemes.
	cachedSchemes map[string]*ColorScheme

	// colorSupport is the detected terminal color support.
	colorSupport *ColorSupport
}

// NewDefaultColorSchemeHandler creates a new DefaultColorSchemeHandler.
func NewDefaultColorSchemeHandler(subsheller SubshellSpawner, injector EnvironmentInjector) (*DefaultColorSchemeHandler, error) {
	handler := &DefaultColorSchemeHandler{
		subsheller:    subsheller,
		injector:      injector,
		cachedSchemes: make(map[string]*ColorScheme),
	}

	// Detect color support
	handler.colorSupport = handler.detectColorSupport()

	return handler, nil
}

// PreviewColorScheme generates a preview for a color scheme.
func (c *DefaultColorSchemeHandler) PreviewColorScheme(ctx context.Context, config *ColorSchemePreviewConfig) (*PreviewResult, error) {
	// Set default timeout
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultColorTimeout
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build environment
	env := c.buildColorEnv(config)

	// Generate preview output
	var sampleOutput string
	if config.ShowPalette {
		sampleOutput = c.generatePalettePreview()
	} else {
		sampleOutput = c.generateSamplePreview(config.ColorValues)
	}

	// Create subshell config
	subshellConfig := &SubshellConfig{
		ShellType:     config.Shell,
		Environment:   env,
		Timeout:       timeout,
		CaptureStdout: true,
		CaptureStderr: true,
		Command:       fmt.Sprintf("echo '%s'", escapeForShell(sampleOutput)),
	}

	// Spawn subshell
	result, err := c.subsheller.Spawn(ctx, subshellConfig)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrColorPreviewFailed, err)
	}

	// Build preview result
	previewResult := &PreviewResult{
		ID:          generatePreviewID(),
		Status:      StatusCompleted,
		Output:      result.Stdout,
		ErrorOutput: result.Stderr,
		ExitCode:    result.ExitCode,
		Duration:    result.Duration,
		Config: PreviewConfig{
			Type:        PreviewTypeColorScheme,
			Shell:       config.Shell,
			ColorScheme: config.ColorSchemeName,
			Timeout:     config.Timeout,
		},
	}

	if result.ExitCode != 0 {
		previewResult.Status = StatusFailed
		previewResult.ErrorMessage = result.Stderr
	}

	return previewResult, nil
}

// GetTerminalColorCapabilities returns the terminal's color support.
func (c *DefaultColorSchemeHandler) GetTerminalColorCapabilities() (supportsTrueColor bool, supports256Color bool, err error) {
	if c.colorSupport == nil {
		c.colorSupport = c.detectColorSupport()
	}

	return c.colorSupport.SupportsTrueColor, c.colorSupport.Supports256, nil
}

// ListAvailableColorSchemes returns a list of built-in color schemes.
func (c *DefaultColorSchemeHandler) ListAvailableColorSchemes() ([]string, error) {
	// Built-in color schemes
	schemes := []string{
		"dracula",
		"nord",
		"gruvbox",
		"solarized-dark",
		"solarized-light",
		"catppuccin",
		"tokyo-night",
		"one-dark",
		"atom-one-light",
		"gotham",
		"monokai",
		"default-dark",
		"default-light",
	}

	return schemes, nil
}

// GetColorScheme returns a color scheme by name.
func (c *DefaultColorSchemeHandler) GetColorScheme(name string) (*ColorScheme, error) {
	// Check cache first
	if scheme, exists := c.cachedSchemes[name]; exists {
		return scheme, nil
	}

	// Load built-in scheme
	scheme, err := c.loadBuiltInScheme(name)
	if err != nil {
		return nil, err
	}

	c.cachedSchemes[name] = scheme
	return scheme, nil
}

// detectColorSupport detects terminal color capabilities.
func (c *DefaultColorSchemeHandler) detectColorSupport() *ColorSupport {
	support := &ColorSupport{
		SupportsBasic:     true, // Assume basic support
		Supports256:       false,
		SupportsTrueColor: false,
	}

	// Check COLORTERM
	colorTerm := os.Getenv("COLORTERM")
	if strings.Contains(strings.ToLower(colorTerm), "truecolor") ||
		strings.Contains(strings.ToLower(colorTerm), "24bit") {
		support.SupportsTrueColor = true
	}

	// Check TERM
	term := os.Getenv("TERM")
	termLower := strings.ToLower(term)

	// 256-color terminals
	if strings.Contains(termLower, "256") ||
		strings.Contains(termLower, "xterm") ||
		strings.Contains(termLower, "screen") ||
		strings.Contains(termLower, "tmux") {
		support.Supports256 = true
	}

	// Known true color terminals
	trueColorTerminals := []string{
		"iterm", "alacritty", "kitty", "wezterm", "foot",
		"contour", "rio", "warp", "ghostty",
	}

	termProgram := os.Getenv("TERM_PROGRAM")
	termProgramLower := strings.ToLower(termProgram)
	for _, tc := range trueColorTerminals {
		if strings.Contains(termLower, tc) || strings.Contains(termProgramLower, tc) {
			support.SupportsTrueColor = true
			support.Supports256 = true
			break
		}
	}

	return support
}

// buildColorEnv builds environment variables for color scheme preview.
func (c *DefaultColorSchemeHandler) buildColorEnv(config *ColorSchemePreviewConfig) map[string]string {
	env := make(map[string]string)

	// Preserve current environment
	for _, pair := range os.Environ() {
		if idx := strings.Index(pair, "="); idx > 0 {
			env[pair[:idx]] = pair[idx+1:]
		}
	}

	// Add color environment
	colorEnv := c.injector.InjectColorSchemeEnv(config.ColorSchemeName)
	for k, v := range colorEnv {
		env[k] = v
	}

	// Force terminal to support colors
	env[EnvTerm] = "xterm-256color"
	env[EnvColorterm] = "truecolor"

	return env
}

// generatePalettePreview generates a color palette preview.
func (c *DefaultColorSchemeHandler) generatePalettePreview() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("Color Palette Preview:\n")
	sb.WriteString("=====================\n\n")

	// Standard 16 colors
	sb.WriteString("Standard Colors:\n")
	colors := []string{"Black", "Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White"}
	for i, color := range colors {
		sb.WriteString(fmt.Sprintf("\033[%dm%-8s\033[0m", 30+i, color))
	}
	sb.WriteString("\n")

	// Bright colors
	sb.WriteString("Bright Colors:\n")
	for i, color := range colors {
		sb.WriteString(fmt.Sprintf("\033[%dm%-8s\033[0m", 90+i, "Bright"+color[:4]))
	}
	sb.WriteString("\n\n")

	// 256-color cube
	sb.WriteString("256-Color Cube (partial):\n")
	for i := 0; i < 16; i++ {
		sb.WriteString(fmt.Sprintf("\033[48;5;%dm  \033[0m", i))
	}
	sb.WriteString("\n")

	// True color test
	sb.WriteString("\nTrue Color Test (if supported):\n")
	sb.WriteString("\033[48;2;255;0;0m Red \033[0m")
	sb.WriteString("\033[48;2;0;255;0m Green \033[0m")
	sb.WriteString("\033[48;2;0;0;255m Blue \033[0m\n")

	return sb.String()
}

// generateSamplePreview generates a sample output preview.
func (c *DefaultColorSchemeHandler) generateSamplePreview(colors map[string]string) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("Sample Terminal Output:\n")
	sb.WriteString("=======================\n\n")

	if colors != nil {
		// Use provided colors
		sb.WriteString(c.colorize(colors, "foreground", "background", "Normal text output") + "\n")
		sb.WriteString(c.colorize(colors, "green", "background", "Success: Operation completed") + "\n")
		sb.WriteString(c.colorize(colors, "yellow", "background", "Warning: Please review") + "\n")
		sb.WriteString(c.colorize(colors, "red", "background", "Error: Something went wrong") + "\n")
		sb.WriteString(c.colorize(colors, "cyan", "background", "Info: Additional details") + "\n")
	} else {
		// Default colors
		sb.WriteString("\033[0mNormal text output\033[0m\n")
		sb.WriteString("\033[32mSuccess: Operation completed\033[0m\n")
		sb.WriteString("\033[33mWarning: Please review\033[0m\n")
		sb.WriteString("\033[31mError: Something went wrong\033[0m\n")
		sb.WriteString("\033[36mInfo: Additional details\033[0m\n")
	}

	sb.WriteString("\n")
	sb.WriteString("$ ls -la\n")
	sb.WriteString("drwxr-xr-x  2 user user 4096 Mar 18 12:00 \033[34m.\033[0m\n")
	sb.WriteString("drwxr-xr-x 15 user user 4096 Mar 18 11:30 \033[34m..\033[0m\n")
	sb.WriteString("-rw-r--r--  1 user user  123 Mar 18 12:00 \033[32mconfig.json\033[0m\n")
	sb.WriteString("-rwxr-xr-x  1 user user 4567 Mar 18 12:00 \033[32mscript.sh\033[0m\n")

	return sb.String()
}

// colorize wraps text with ANSI color codes.
func (c *DefaultColorSchemeHandler) colorize(colors map[string]string, fg, bg, text string) string {
	fgColor := colors[fg]
	bgColor := colors[bg]

	if fgColor != "" && bgColor != "" {
		return fmt.Sprintf("\033[38;2;%sm\033[48;2;%sm%s\033[0m", fgColor, bgColor, text)
	} else if fgColor != "" {
		return fmt.Sprintf("\033[38;2;%sm%s\033[0m", fgColor, text)
	}
	return text
}

// loadBuiltInScheme loads a built-in color scheme.
func (c *DefaultColorSchemeHandler) loadBuiltInScheme(name string) (*ColorScheme, error) {
	// Built-in color schemes
	schemes := map[string]*ColorScheme{
		"dracula": {
			Name:        "dracula",
			DisplayName: "Dracula",
			Description: "Dark theme with purple accents",
			Colors: ColorDefinitions{
				Foreground: "#f8f8f2",
				Background: "#282a36",
				Normal: ColorPalette{
					Black: "#000000", Red: "#ff5555", Green: "#50fa7b", Yellow: "#f1fa8c",
					Blue: "#bd93f9", Magenta: "#ff79c6", Cyan: "#8be9fd", White: "#bfbfbf",
				},
				Bright: ColorPalette{
					Black: "#282a36", Red: "#ff5555", Green: "#50fa7b", Yellow: "#f1fa8c",
					Blue: "#bd93f9", Magenta: "#ff79c6", Cyan: "#8be9fd", White: "#f8f8f2",
				},
			},
		},
		"nord": {
			Name:        "nord",
			DisplayName: "Nord",
			Description: "Arctic, bluish color palette",
			Colors: ColorDefinitions{
				Foreground: "#d8dee9",
				Background: "#2e3440",
				Normal: ColorPalette{
					Black: "#3b4252", Red: "#bf616a", Green: "#a3be8c", Yellow: "#ebcb8b",
					Blue: "#81a1c1", Magenta: "#b48ead", Cyan: "#88c0d0", White: "#e5e9f0",
				},
				Bright: ColorPalette{
					Black: "#4c566a", Red: "#bf616a", Green: "#a3be8c", Yellow: "#ebcb8b",
					Blue: "#81a1c1", Magenta: "#b48ead", Cyan: "#8be9fd", White: "#eceff4",
				},
			},
		},
		"gruvbox": {
			Name:        "gruvbox",
			DisplayName: "Gruvbox",
			Description: "Retro groove color palette",
			Colors: ColorDefinitions{
				Foreground: "#ebdbb2",
				Background: "#282828",
				Normal: ColorPalette{
					Black: "#282828", Red: "#cc241d", Green: "#98971a", Yellow: "#d79921",
					Blue: "#458588", Magenta: "#b16286", Cyan: "#689d6a", White: "#a89984",
				},
				Bright: ColorPalette{
					Black: "#928374", Red: "#fb4934", Green: "#b8bb26", Yellow: "#fabd2f",
					Blue: "#83a598", Magenta: "#d3869b", Cyan: "#8ec07c", White: "#ebdbb2",
				},
			},
		},
		"solarized-dark": {
			Name:        "solarized-dark",
			DisplayName: "Solarized Dark",
			Description: "Precision color palette for machines and people",
			Colors: ColorDefinitions{
				Foreground: "#839496",
				Background: "#002b36",
				Normal: ColorPalette{
					Black: "#073642", Red: "#dc322f", Green: "#859900", Yellow: "#b58900",
					Blue: "#268bd2", Magenta: "#d33682", Cyan: "#2aa198", White: "#eee8d5",
				},
				Bright: ColorPalette{
					Black: "#002b36", Red: "#cb4b16", Green: "#586e75", Yellow: "#657b83",
					Blue: "#839496", Magenta: "#6c71c4", Cyan: "#93a1a1", White: "#fdf6e3",
				},
			},
		},
	}

	scheme, exists := schemes[name]
	if !exists {
		return nil, ErrColorSchemeNotFound
	}

	return scheme, nil
}

// ValidateColorScheme validates a color scheme definition.
func ValidateColorScheme(scheme *ColorScheme) error {
	if scheme.Name == "" {
		return errors.New("color scheme name is required")
	}

	if scheme.Colors.Foreground == "" {
		return errors.New("foreground color is required")
	}

	if scheme.Colors.Background == "" {
		return errors.New("background color is required")
	}

	return nil
}

// GetDefaultColorSchemes returns the default color schemes.
func GetDefaultColorSchemes() []*ColorScheme {
	return []*ColorScheme{
		{Name: "dracula", DisplayName: "Dracula"},
		{Name: "nord", DisplayName: "Nord"},
		{Name: "gruvbox", DisplayName: "Gruvbox"},
		{Name: "solarized-dark", DisplayName: "Solarized Dark"},
		{Name: "solarized-light", DisplayName: "Solarized Light"},
		{Name: "catppuccin", DisplayName: "Catppuccin"},
		{Name: "tokyo-night", DisplayName: "Tokyo Night"},
		{Name: "one-dark", DisplayName: "One Dark"},
	}
}
