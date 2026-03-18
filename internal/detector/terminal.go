// Package detector provides system detection capabilities.
// This file implements terminal emulator detection.
package detector

import (
	"os"
	"strings"
)

// terminalDetector implements TerminalDetector interface.
type terminalDetector struct{}

// NewTerminalDetector creates a new terminal detector.
func NewTerminalDetector() TerminalDetector {
	return &terminalDetector{}
}

// Detect implements TerminalDetector.Detect.
func (d *terminalDetector) Detect() (*TerminalInfo, error) {
	info := &TerminalInfo{
		Type:                  TerminalTypeUnknown,
		Name:                  "unknown",
		SupportsTrueColor:     false,
		SupportsLigatures:     false,
		SupportsHyperlinks:    false,
		SupportsKittyGraphics: false,
	}

	// Detect terminal from environment variables
	info.Type, info.Name = d.detectTerminalType()
	info.Version = d.detectTerminalVersion(info.Type)

	// Detect capabilities
	info.SupportsTrueColor = d.detectTrueColorSupport()
	info.SupportsLigatures = d.detectLigatureSupport(info.Type)
	info.SupportsHyperlinks = d.detectHyperlinkSupport()
	info.SupportsKittyGraphics = info.Type == TerminalTypeKitty

	// Detect font settings
	info.FontFamily, info.FontSize = d.detectFontSettings(info.Type)

	return info, nil
}

// detectTerminalType determines the terminal emulator type.
func (d *terminalDetector) detectTerminalType() (TerminalType, string) {
	// Check for iTerm2 on macOS
	if termProgram := os.Getenv("TERM_PROGRAM"); termProgram != "" {
		switch termProgram {
		case "iTerm.app":
			return TerminalTypeITerm2, "iTerm2"
		case "Apple_Terminal":
			return TerminalTypeUnknown, "Apple Terminal"
		case "vscode":
			return TerminalTypeVSCode, "VS Code"
		case "WezTerm":
			return TerminalTypeWezTerm, "WezTerm"
		}
	}

	// Check for WezTerm (alternative detection)
	if os.Getenv("WEZTERM_EXECUTABLE") != "" || os.Getenv("WEZTERM_PANE") != "" {
		return TerminalTypeWezTerm, "WezTerm"
	}

	// Check for Alacritty
	if os.Getenv("ALACRITTY_WINDOW_ID") != "" {
		return TerminalTypeAlacritty, "Alacritty"
	}

	// Check for Kitty
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return TerminalTypeKitty, "Kitty"
	}

	// Check for Windows Terminal
	if os.Getenv("WT_SESSION") != "" {
		return TerminalTypeWindowsTerminal, "Windows Terminal"
	}

	// Check for GNOME Terminal
	if term := os.Getenv("GNOME_TERMINAL_SCREEN"); term != "" {
		return TerminalTypeGNOMETerminal, "GNOME Terminal"
	}

	// Check for Konsole
	if term := os.Getenv("KONSOLE_VERSION"); term != "" {
		return TerminalTypeKonsole, "Konsole"
	}

	// Check TERM_PROGRAM_VERSION for VS Code
	if os.Getenv("TERM_PROGRAM_VERSION") != "" {
		return TerminalTypeVSCode, "VS Code"
	}

	// Fallback to TERM variable
	term := os.Getenv("TERM")
	if strings.Contains(term, "xterm") || strings.Contains(term, "screen") {
		return TerminalTypeUnknown, "xterm-compatible"
	}

	return TerminalTypeUnknown, "unknown"
}

// detectTerminalVersion returns the terminal version if available.
func (d *terminalDetector) detectTerminalVersion(termType TerminalType) string {
	switch termType {
	case TerminalTypeITerm2:
		return os.Getenv("TERM_PROGRAM_VERSION")
	case TerminalTypeVSCode:
		return os.Getenv("TERM_PROGRAM_VERSION")
	case TerminalTypeKitty:
		if version := os.Getenv("KITTY_VERSION"); version != "" {
			return version
		}
	case TerminalTypeWezTerm:
		if version := os.Getenv("WEZTERM_VERSION"); version != "" {
			return version
		}
	}
	return ""
}

// detectTrueColorSupport checks if the terminal supports true color (24-bit).
func (d *terminalDetector) detectTrueColorSupport() bool {
	// Check COLORTERM environment variable
	colorterm := os.Getenv("COLORTERM")
	if colorterm == "truecolor" || colorterm == "24bit" {
		return true
	}

	// Check TERM variable for true color support
	term := os.Getenv("TERM")
	if strings.Contains(term, "truecolor") || strings.Contains(term, "24bit") {
		return true
	}

	// Known terminals that support true color
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "iTerm.app":
		return true
	case "WezTerm":
		return true
	}

	if os.Getenv("ALACRITTY_WINDOW_ID") != "" {
		return true
	}

	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}

	if os.Getenv("WT_SESSION") != "" {
		return true
	}

	return false
}

// detectLigatureSupport checks if the terminal supports font ligatures.
func (d *terminalDetector) detectLigatureSupport(termType TerminalType) bool {
	// Terminals known to support ligatures
	switch termType {
	case TerminalTypeITerm2, TerminalTypeAlacritty, TerminalTypeKitty, TerminalTypeVSCode, TerminalTypeWezTerm:
		return true
	default:
		return false
	}
}

// detectHyperlinkSupport checks if the terminal supports OSC 8 hyperlinks.
func (d *terminalDetector) detectHyperlinkSupport() bool {
	// Most modern terminals support OSC 8 hyperlinks
	termProgram := os.Getenv("TERM_PROGRAM")
	switch termProgram {
	case "iTerm.app":
		return true
	case "WezTerm":
		return true
	}

	// Check if terminal supports the capability
	term := os.Getenv("TERM")
	if strings.Contains(term, "truecolor") {
		return true
	}

	return os.Getenv("ALACRITTY_WINDOW_ID") != "" ||
		os.Getenv("KITTY_WINDOW_ID") != "" ||
		os.Getenv("WT_SESSION") != "" ||
		os.Getenv("WEZTERM_EXECUTABLE") != ""
}

// detectFontSettings attempts to detect terminal font family and size.
// This is limited to terminals that expose this information via environment variables.
func (d *terminalDetector) detectFontSettings(termType TerminalType) (string, int) {
	// WezTerm exposes font information
	if fontFamily := os.Getenv("WEZTERM_FONT_FAMILY"); fontFamily != "" {
		fontSize := 0
		if size := os.Getenv("WEZTERM_FONT_SIZE"); size != "" {
			// Parse font size if available
			if n, err := parseInt(size); err == nil {
				fontSize = n
			}
		}
		return fontFamily, fontSize
	}

	// iTerm2: Font settings can be queried but require special handling
	// For now, we'll return empty values as iTerm2 doesn't expose fonts via env vars

	// Kitty: Can query via kitty @ get-fonts remote control
	// This requires kitty remote control which is complex for detection

	// Alacritty: Font is in config file, not exposed via env vars
	// Would require parsing config file at ~/.config/alacritty/alacritty.yml

	// VS Code: Font settings are in VS Code settings
	// Would require parsing settings.json

	// Return empty for terminals that don't expose font info
	return "", 0
}

// parseInt is a simple helper for parsing integers.
func parseInt(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}

	var result int
	var negative bool

	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	}

	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		result = result*10 + int(c-'0')
	}

	if negative {
		result = -result
	}

	return result, nil
}
