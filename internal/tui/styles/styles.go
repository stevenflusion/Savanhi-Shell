// Package styles provides Lipgloss styling for the TUI.
package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// ColorPalette defines the color scheme for the TUI.
var ColorPalette = struct {
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Accent     lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
	Muted      lipgloss.Color
	Error      lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
}{
	Primary:    lipgloss.Color("#7C3AED"), // Purple
	Secondary:  lipgloss.Color("#3B82F6"), // Blue
	Accent:     lipgloss.Color("#10B981"), //Green
	Background: lipgloss.Color("#1F2937"), // Dark gray
	Foreground: lipgloss.Color("#F9FAFB"), // Light gray
	Muted:      lipgloss.Color("#6B7280"), // Gray
	Error:      lipgloss.Color("#EF4444"), // Red
	Success:    lipgloss.Color("#10B981"), // Green
	Warning:    lipgloss.Color("#F59E0B"), // Yellow
}

// Common styles.
var (
	// Title is the main title style.
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPalette.Primary).
		MarginBottom(1)

	// Subtitle is the subtitle style.
	Subtitle = lipgloss.NewStyle().
			Foreground(ColorPalette.Secondary).
			MarginBottom(1)

	// Label is the label style.
	Label = lipgloss.NewStyle().
		Foreground(ColorPalette.Muted).
		Width(15).
		Align(lipgloss.Right)

	// Value is the value style.
	Value = lipgloss.NewStyle().
		Foreground(ColorPalette.Foreground).
		Bold(true)

	// Info is informational text style.
	Info = lipgloss.NewStyle().
		Foreground(ColorPalette.Secondary)

	// Success is success text style.
	Success = lipgloss.NewStyle().
		Foreground(ColorPalette.Success).
		Bold(true)

	// Warning is warning text style.
	Warning = lipgloss.NewStyle().
		Foreground(ColorPalette.Warning)

	// Error is error text style.
	Error = lipgloss.NewStyle().
		Foreground(ColorPalette.Error).
		Bold(true)

	// Muted is muted text style.
	Muted = lipgloss.NewStyle().
		Foreground(ColorPalette.Muted)

	// Highlight is highlighted text style.
	Highlight = lipgloss.NewStyle().
			Foreground(ColorPalette.Foreground).
			Background(ColorPalette.Primary).
			Bold(true).
			Padding(0, 1)

	// Border is the border style.
	Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPalette.Primary).
		Padding(0, 1)

	// Box is a bordered box style.
	Box = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorPalette.Muted).
		Padding(1, 2).
		Margin(1, 0)
)

// KeyStyle styles for keybinding hints.
var KeyStyle = struct {
	Key   lipgloss.Style
	Desc  lipgloss.Style
	Group lipgloss.Style
}{
	Key: lipgloss.NewStyle().
		Foreground(ColorPalette.Foreground).
		Background(ColorPalette.Primary).
		Padding(0, 1).
		Bold(true),
	Desc: lipgloss.NewStyle().
		Foreground(ColorPalette.Muted),
	Group: lipgloss.NewStyle().
		MarginTop(1),
}

// ItemStyle styles for list items.
var ItemStyle = struct {
	Normal   lipgloss.Style
	Selected lipgloss.Style
	Active   lipgloss.Style
}{
	Normal: lipgloss.NewStyle().
		Foreground(ColorPalette.Foreground).
		Padding(0, 1),
	Selected: lipgloss.NewStyle().
		Foreground(ColorPalette.Foreground).
		Background(ColorPalette.Primary).
		Bold(true).
		Padding(0, 1),
	Active: lipgloss.NewStyle().
		Foreground(ColorPalette.Accent).
		Background(ColorPalette.Background).
		Bold(true).
		Padding(0, 1),
}

// StatusStyle styles for status displays.
var StatusStyle = struct {
	Ready   lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
}{
	Ready: lipgloss.NewStyle().
		Foreground(ColorPalette.Accent).
		Bold(true),
	Success: lipgloss.NewStyle().
		Foreground(ColorPalette.Success).
		Bold(true),
	Error: lipgloss.NewStyle().
		Foreground(ColorPalette.Error).
		Bold(true),
	Warning: lipgloss.NewStyle().
		Foreground(ColorPalette.Warning).
		Bold(true),
}

// HeaderStyle styles for header area.
var HeaderStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPalette.Primary).
	MarginBottom(1)

// FooterStyle styles for footer area.
var FooterStyle = lipgloss.NewStyle().
	Foreground(ColorPalette.Muted).
	MarginTop(1)

// WelcomeBox is the main welcome box style.
var WelcomeBox = lipgloss.NewStyle().
	Border(lipgloss.DoubleBorder()).
	BorderForeground(ColorPalette.Primary).
	Padding(1, 3).
	Margin(1).
	Width(60).
	Align(lipgloss.Center)

// LogoStyle for the Savanhi logo.
var LogoStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(ColorPalette.Primary).
	Background(lipgloss.Color("#000000")).
	Padding(0, 2)

// ProgressStyle for progress bars.
var ProgressStyle = lipgloss.NewStyle().
	Foreground(ColorPalette.Accent)

// DetectionBox styles for detection results.
var DetectionBox = struct {
	Container lipgloss.Style
	Item      lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
}{
	Container: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPalette.Secondary).
		Padding(1, 2).
		Width(60),
	Item: lipgloss.NewStyle().
		Margin(0, 1),
	Label: lipgloss.NewStyle().
		Foreground(ColorPalette.Muted).
		Width(12).
		Align(lipgloss.Right),
	Value: lipgloss.NewStyle().
		Foreground(ColorPalette.Foreground).
		Bold(true),
}

// HealthBox styles for health dashboard display.
var HealthBox = struct {
	Container    lipgloss.Style
	Section      lipgloss.Style
	SectionTitle lipgloss.Style
	Item         lipgloss.Style
	Label        lipgloss.Style
	Value        lipgloss.Style
	Success      lipgloss.Style
	Error        lipgloss.Style
	Warning      lipgloss.Style
}{
	Container: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPalette.Accent).
		Padding(1, 2).
		Width(80),
	Section: lipgloss.NewStyle().
		Margin(1, 0).
		Padding(0, 1),
	SectionTitle: lipgloss.NewStyle().
		Foreground(ColorPalette.Primary).
		Bold(true).
		MarginBottom(1),
	Item: lipgloss.NewStyle().
		Margin(0, 1),
	Label: lipgloss.NewStyle().
		Foreground(ColorPalette.Muted).
		Width(15).
		Align(lipgloss.Right),
	Value: lipgloss.NewStyle().
		Foreground(ColorPalette.Foreground).
		Bold(true),
	Success: lipgloss.NewStyle().
		Foreground(ColorPalette.Success).
		Bold(true),
	Error: lipgloss.NewStyle().
		Foreground(ColorPalette.Error).
		Bold(true),
	Warning: lipgloss.NewStyle().
		Foreground(ColorPalette.Warning).
		Bold(true),
}

// ButtonStyle for interactive buttons.
var ButtonStyle = struct {
	Normal   lipgloss.Style
	Selected lipgloss.Style
}{
	Normal: lipgloss.NewStyle().
		Foreground(ColorPalette.Foreground).
		Background(ColorPalette.Primary).
		Padding(0, 2).
		Bold(true),
	Selected: lipgloss.NewStyle().
		Foreground(ColorPalette.Primary).
		Background(ColorPalette.Foreground).
		Padding(0, 2).
		Bold(true),
}

// FormatKey formats a keybinding hint.
func FormatKey(key, description string) string {
	return KeyStyle.Key.Render(key) + " " + KeyStyle.Desc.Render(description)
}

// FormatKeyGroup formats a group of keybindings.
func FormatKeyGroup(title string, keys []string) string {
	result := KeyStyle.Group.Render(title) + "\n"
	for _, key := range keys {
		result += "  " + key + "\n"
	}
	return result
}
