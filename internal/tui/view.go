// Package tui provides view rendering for the TUI.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/savanhi/shell/internal/tui/styles"
)

// renderLoading renders the loading screen.
func (m Model) renderLoading() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Savanhi Shell"))
	b.WriteString("\n\n")

	if m.LoadingMessage != "" {
		b.WriteString(styles.Info.Render(m.LoadingMessage))
		b.WriteString("\n")
	} else {
		b.WriteString(styles.Info.Render("Loading..."))
		b.WriteString("\n")
	}

	return b.String()
}

// renderWelcome renders the welcome screen.
func (m Model) renderWelcome() string {
	var b strings.Builder

	// Logo/Title
	title := styles.LogoStyle.Render("  Savanhi Shell  ")
	b.WriteString("\n")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Welcome message
	welcome := styles.WelcomeBox.Render(
		styles.Title.Render("Welcome to Savanhi Shell!") + "\n\n" +
			styles.Info.Render("Configure your shell environment") + "\n" +
			styles.Info.Render("with themes, fonts, and productivity tools.") + "\n\n" +
			styles.Muted.Render("Press Enter to begin..."),
	)
	b.WriteString(welcome)
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderDetect renders the system detection screen.
func (m Model) renderDetect() string {
	var b strings.Builder

	// Title
	b.WriteString("\n")
	b.WriteString(styles.Title.Render("System Detection"))
	b.WriteString("\n\n")

	// Detection results box
	detectBox := styles.DetectionBox.Container

	var content strings.Builder

	if m.systemInfo != nil {
		// OS
		content.WriteString(styles.DetectionBox.Label.Render("OS:"))
		content.WriteString(" ")
		if m.systemInfo.OS != "" {
			content.WriteString(styles.DetectionBox.Value.Render(m.systemInfo.OS))
		} else {
			content.WriteString(styles.Warning.Render("Not detected"))
		}
		content.WriteString("\n")

		// Shell
		content.WriteString(styles.DetectionBox.Label.Render("Shell:"))
		content.WriteString(" ")
		if m.systemInfo.Shell != "" {
			content.WriteString(styles.DetectionBox.Value.Render(m.systemInfo.Shell))
		} else {
			content.WriteString(styles.Warning.Render("Not detected"))
		}
		content.WriteString("\n")

		// Terminal
		content.WriteString(styles.DetectionBox.Label.Render("Terminal:"))
		content.WriteString(" ")
		if m.systemInfo.Terminal != "" {
			content.WriteString(styles.DetectionBox.Value.Render(m.systemInfo.Terminal))
		} else {
			content.WriteString(styles.Warning.Render("Not detected"))
		}
		content.WriteString("\n")

		// Fonts
		content.WriteString(styles.DetectionBox.Label.Render("Fonts:"))
		content.WriteString(" ")
		if len(m.systemInfo.Fonts) > 0 {
			fontList := strings.Join(m.systemInfo.Fonts, ", ")
			if len(fontList) > 50 {
				fontList = fontList[:50] + "..."
			}
			content.WriteString(styles.DetectionBox.Value.Render(fontList))
		} else {
			content.WriteString(styles.Warning.Render("No Nerd Fonts detected"))
		}
		content.WriteString("\n")

		// Show detailed info from DetectorResult
		if m.DetectorResult != nil {
			content.WriteString("\n")
			content.WriteString(styles.Subtitle.Render("Installed Components:"))
			content.WriteString("\n")

			// Check for oh-my-posh
			if m.DetectorResult.ExistingConfigs != nil && m.DetectorResult.ExistingConfigs.HasOhMyPosh {
				content.WriteString(styles.StatusStyle.Success.Render("✓ oh-my-posh installed"))
				content.WriteString("\n")
			}

			// Check for starship
			if m.DetectorResult.ExistingConfigs != nil && m.DetectorResult.ExistingConfigs.HasStarship {
				content.WriteString(styles.StatusStyle.Success.Render("✓ starship installed"))
				content.WriteString("\n")
			}
		}
	} else {
		content.WriteString(styles.Info.Render("Detecting system information..."))
		content.WriteString("\n")
	}

	b.WriteString(detectBox.Render(content.String()))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderThemeSelect renders the theme selection screen.
func (m Model) renderThemeSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Select Theme"))
	b.WriteString("\n\n")

	// Render theme list
	for i, item := range m.Items {
		if i == m.Cursor {
			b.WriteString(styles.ItemStyle.Selected.Render("→ " + item))
		} else {
			b.WriteString(styles.ItemStyle.Normal.Render("  " + item))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderFontSelect renders the font selection screen.
func (m Model) renderFontSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Select Font"))
	b.WriteString("\n\n")

	// Render font list
	for i, item := range m.Items {
		if i == m.Cursor {
			b.WriteString(styles.ItemStyle.Selected.Render("→ " + item))
		} else {
			b.WriteString(styles.ItemStyle.Normal.Render("  " + item))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderPreview renders the preview screen.
func (m Model) renderPreview() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Preview"))
	b.WriteString("\n\n")

	// Build preview content
	var content strings.Builder

	// Show selected theme
	content.WriteString(styles.Subtitle.Render("Selected Configuration"))
	content.WriteString("\n\n")

	// Theme
	content.WriteString(styles.DetectionBox.Label.Render("Theme:"))
	content.WriteString(" ")
	if m.SelectedTheme != "" {
		content.WriteString(styles.DetectionBox.Value.Render(m.SelectedTheme))
	} else {
		content.WriteString(styles.Warning.Render("None selected"))
	}
	content.WriteString("\n")

	// Font
	content.WriteString(styles.DetectionBox.Label.Render("Font:"))
	content.WriteString(" ")
	if m.SelectedFont != "" {
		content.WriteString(styles.DetectionBox.Value.Render(m.SelectedFont))
	} else {
		content.WriteString(styles.Warning.Render("None selected"))
	}
	content.WriteString("\n")

	// Shell info
	if m.systemInfo != nil {
		content.WriteString("\n")
		content.WriteString(styles.Subtitle.Render("System"))
		content.WriteString("\n")
		content.WriteString(styles.DetectionBox.Label.Render("Shell:"))
		content.WriteString(" ")
		content.WriteString(styles.DetectionBox.Value.Render(m.systemInfo.Shell))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(styles.Muted.Render("Press Enter to install or Esc to go back"))

	b.WriteString(styles.Box.Render(content.String()))
	b.WriteString("\n\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderInstall renders the installation screen.
func (m Model) renderInstall() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Installing"))
	b.WriteString("\n\n")

	if m.Loading {
		b.WriteString(styles.Info.Render(m.LoadingMessage))
	} else {
		b.WriteString(styles.Success.Render("✓ Installation complete!"))
	}

	b.WriteString("\n\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderComplete renders the completion screen.
func (m Model) renderComplete() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Complete!"))
	b.WriteString("\n\n")

	b.WriteString(styles.Success.Render("Your shell has been configured successfully!"))
	b.WriteString("\n\n")

	if m.systemInfo != nil {
		b.WriteString(styles.Info.Render("Restart your shell to see changes."))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderError renders the error screen.
func (m Model) renderError() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Error"))
	b.WriteString("\n\n")

	if m.Error != nil {
		b.WriteString(styles.Error.Render(m.Error.Error()))
	} else {
		b.WriteString(styles.Error.Render("An unknown error occurred"))
	}

	b.WriteString("\n\n")
	b.WriteString(styles.Info.Render("Press Esc to go back or q to quit"))
	b.WriteString("\n")

	return b.String()
}

// renderFooter renders the footer with keybindings.
func (m Model) renderFooter() string {
	var b strings.Builder

	b.WriteString(styles.FooterStyle.Render("─"))
	b.WriteString("\n")

	// Build key hints based on current screen
	var keys []string

	switch m.CurrentScreen {
	case ScreenWelcome:
		keys = append(keys, styles.FormatKey("Enter", "Continue"))
	case ScreenDetect:
		keys = append(keys, styles.FormatKey("Enter", "Continue"))
		keys = append(keys, styles.FormatKey("H", "Health"))
		keys = append(keys, styles.FormatKey("Esc", "Back"))
	case ScreenPluginSelect:
		keys = append(keys, styles.FormatKey("↑↓/jk", "Navigate"))
		keys = append(keys, styles.FormatKey("Space", "Toggle"))
		keys = append(keys, styles.FormatKey("Enter", "Continue"))
		keys = append(keys, styles.FormatKey("Esc", "Back"))
	case ScreenThemeSelect, ScreenFontSelect:
		keys = append(keys, styles.FormatKey("↑↓", "Navigate"))
		keys = append(keys, styles.FormatKey("Enter", "Select"))
		keys = append(keys, styles.FormatKey("Esc", "Back"))
	case ScreenPreview:
		keys = append(keys, styles.FormatKey("Enter", "Install"))
		keys = append(keys, styles.FormatKey("Esc", "Back"))
	case ScreenInstall:
		if m.Loading {
			keys = append(keys, styles.FormatKey("Esc", "Cancel"))
		} else {
			keys = append(keys, styles.FormatKey("Enter", "Continue"))
		}
	case ScreenComplete:
		keys = append(keys, styles.FormatKey("Enter", "Finish"))
	case ScreenError:
		keys = append(keys, styles.FormatKey("Esc", "Back"))
	}

	keys = append(keys, styles.FormatKey("q", "Quit"))

	b.WriteString(strings.Join(keys, "  "))
	b.WriteString("\n")

	return b.String()
}

// JoinHorizontal joins strings horizontally with lipgloss layout.
func JoinHorizontal(strs ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, strs...)
}

// JoinVertical joins strings vertically with lipgloss layout.
func JoinVertical(strs ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, strs...)
}

// FormatStatus formats a status line with label and value.
func FormatStatus(label, value string, success bool) string {
	labelStyle := styles.DetectionBox.Label
	var valueStyle lipgloss.Style
	if success {
		valueStyle = styles.StatusStyle.Success
	} else {
		valueStyle = styles.StatusStyle.Warning
	}

	return fmt.Sprintf("%s %s",
		labelStyle.Render(label),
		valueStyle.Render(value),
	)
}

// renderHealthDashboard renders the health dashboard screen.
func (m Model) renderHealthDashboard() string {
	var b strings.Builder

	// Title with help hint
	titleLine := styles.Title.Render("Terminal Health Dashboard")
	helpHint := styles.Muted.Render("[?]Help")
	b.WriteString("\n")
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, titleLine, "   ", helpHint))
	b.WriteString("\n\n")

	// Check if health data is available
	if m.HealthData == nil {
		b.WriteString(styles.Info.Render("Loading health information..."))
		b.WriteString("\n\n")
		b.WriteString(m.renderHealthFooter())
		return b.String()
	}

	// Section 1: Terminal Capabilities
	b.WriteString(m.renderTerminalCapabilitiesSection())
	b.WriteString("\n")

	// Section 2: Installed Components
	b.WriteString(m.renderInstalledComponentsSection())
	b.WriteString("\n")

	// Section 3: Font Test
	b.WriteString(m.renderFontTestSection())
	b.WriteString("\n")

	// Section 4: Color Test
	b.WriteString(m.renderColorTestSection())
	b.WriteString("\n")

	// Footer with keybindings
	b.WriteString(m.renderHealthFooter())

	return b.String()
}

// renderTerminalCapabilitiesSection renders the terminal capabilities section.
func (m Model) renderTerminalCapabilitiesSection() string {
	var content strings.Builder

	content.WriteString(styles.HealthBox.SectionTitle.Render("TERMINAL CAPABILITIES"))
	content.WriteString("\n")

	if m.HealthData.Terminal == nil {
		content.WriteString(styles.Warning.Render("  Not detected"))
		return styles.HealthBox.Container.Render(content.String())
	}

	caps := m.HealthData.Terminal

	// True Color
	content.WriteString(formatCapabilityLine("True Color", caps.TrueColor, "24-bit color support"))
	content.WriteString("\n")

	// Ligatures
	content.WriteString(formatCapabilityLine("Ligatures", caps.Ligatures, "Font ligatures enabled"))
	content.WriteString("\n")

	// Hyperlinks
	content.WriteString(formatCapabilityLine("Hyperlinks", caps.Hyperlinks, "OSC8 hyperlinks"))
	content.WriteString("\n")

	// Kitty Graphics
	content.WriteString(formatCapabilityLine("Kitty Graphics", caps.KittyGraphics, "Kitty image protocol"))

	return styles.HealthBox.Container.Render(content.String())
}

// renderInstalledComponentsSection renders the installed components section.
func (m Model) renderInstalledComponentsSection() string {
	var content strings.Builder

	content.WriteString(styles.HealthBox.SectionTitle.Render("INSTALLED COMPONENTS"))
	content.WriteString("\n")

	if m.HealthData.Components == nil || len(m.HealthData.Components) == 0 {
		content.WriteString(styles.Warning.Render("  No components detected"))
		return styles.HealthBox.Container.Render(content.String())
	}

	// Display components in a consistent order
	componentOrder := []string{"oh-my-posh", "fonts", "zoxide", "fzf", "bat", "eza"}

	displayedCount := 0
	for _, name := range componentOrder {
		if status, ok := m.HealthData.Components[name]; ok {
			content.WriteString(formatComponentLine(status))
			content.WriteString("\n")
			displayedCount++
		}
	}

	// Display any remaining components not in the ordered list
	for name, status := range m.HealthData.Components {
		found := false
		for _, orderedName := range componentOrder {
			if orderedName == name {
				found = true
				break
			}
		}
		if !found {
			content.WriteString(formatComponentLine(status))
			content.WriteString("\n")
			displayedCount++
		}
	}

	if displayedCount == 0 {
		content.WriteString(styles.Warning.Render("  No components detected"))
	}

	return styles.HealthBox.Container.Render(content.String())
}

// renderFontTestSection renders the font test section.
func (m Model) renderFontTestSection() string {
	var content strings.Builder

	content.WriteString(styles.HealthBox.SectionTitle.Render("FONT TEST"))
	content.WriteString("\n")

	if m.HealthData.FontTest == nil {
		content.WriteString(styles.Warning.Render("  Font test not available"))
		return styles.HealthBox.Container.Render(content.String())
	}

	fontTest := m.HealthData.FontTest

	// Display test glyphs
	var glyphDisplay strings.Builder
	if fontTest.GlyphsRendered {
		// Show a few representative glyphs
		glyphs := fontTest.TestGlyphs
		if len(glyphs) >= 4 {
			glyphDisplay.WriteString(" ")
			glyphDisplay.WriteString(glyphs[0]) // folder
			glyphDisplay.WriteString("  ")
			glyphDisplay.WriteString(glyphs[1]) // file
			glyphDisplay.WriteString("  ")
			glyphDisplay.WriteString(glyphs[10]) // check
			glyphDisplay.WriteString("  ")
			glyphDisplay.WriteString(glyphs[11]) // cross
			glyphDisplay.WriteString(" ")
		}

		content.WriteString(styles.HealthBox.Item.Render(
			styles.StatusStyle.Success.Render(glyphDisplay.String()) + "   All glyphs render correctly",
		))
	} else {
		content.WriteString(styles.HealthBox.Item.Render(
			styles.StatusStyle.Warning.Render("[ASCII fallback]") + "   Some glyphs may not render",
		))
	}

	return styles.HealthBox.Container.Render(content.String())
}

// renderColorTestSection renders the color test section.
func (m Model) renderColorTestSection() string {
	var content strings.Builder

	content.WriteString(styles.HealthBox.SectionTitle.Render("COLOR TEST"))
	content.WriteString("\n")

	if m.HealthData.ColorTest == nil {
		content.WriteString(styles.Warning.Render("  Color test not available"))
		return styles.HealthBox.Container.Render(content.String())
	}

	colorTest := m.HealthData.ColorTest

	// Display color mode
	var modeDesc string
	switch colorTest.ColorMode {
	case "truecolor":
		modeDesc = "True color (24-bit)"
	case "256":
		modeDesc = "256-color mode"
	default:
		modeDesc = "ANSI 16-color mode"
	}

	// Generate a simple color gradient representation
	gradient := generateColorGradient(colorTest.ColorMode)
	content.WriteString(styles.HealthBox.Item.Render(gradient + "  " + modeDesc))

	return styles.HealthBox.Container.Render(content.String())
}

// renderHealthFooter renders the footer with keybindings for the health dashboard.
func (m Model) renderHealthFooter() string {
	var b strings.Builder

	b.WriteString(styles.FooterStyle.Render("─"))
	b.WriteString("\n")

	// Key hints
	keys := []string{
		styles.FormatKey("R", "Refresh"),
		styles.FormatKey("E", "Export"),
		styles.FormatKey("Q", "Quit"),
	}

	b.WriteString(strings.Join(keys, "  "))
	b.WriteString("\n")

	return b.String()
}

// formatCapabilityLine formats a single capability line with status.
func formatCapabilityLine(name string, enabled bool, description string) string {
	var statusIcon string
	var statusStyle lipgloss.Style

	if enabled {
		statusIcon = "✓"
		statusStyle = styles.StatusStyle.Success
	} else {
		statusIcon = "✗"
		statusStyle = styles.StatusStyle.Error
	}

	// Format: "  ✓ True Color      24-bit color support"
	return fmt.Sprintf("  %s %-15s %s",
		statusStyle.Render(statusIcon),
		styles.HealthBox.Value.Render(name),
		styles.HealthBox.Label.Render(description),
	)
}

// formatComponentLine formats a single component status line.
func formatComponentLine(status *ComponentStatus) string {
	if status == nil {
		return ""
	}

	var statusIcon string
	var statusStyle lipgloss.Style

	if status.Installed {
		statusIcon = "✓"
		statusStyle = styles.StatusStyle.Success
	} else {
		statusIcon = "✗"
		statusStyle = styles.StatusStyle.Error
	}

	// Format: "  ✓ oh-my-posh     v19.2.0  /usr/local/bin/oh-my-posh"
	// Or:     "  ✗ bat            NOT INSTALLED"
	var line strings.Builder
	line.WriteString("  ")
	line.WriteString(statusStyle.Render(statusIcon))
	line.WriteString(" ")

	// Name (padded)
	name := status.Name
	if len(name) > 15 {
		name = name[:15]
	}
	line.WriteString(fmt.Sprintf("%-15s", styles.HealthBox.Value.Render(name)))

	if status.Installed && status.Version != "" {
		line.WriteString(" ")
		line.WriteString(styles.HealthBox.Label.Render(status.Version))
	} else if !status.Installed {
		line.WriteString(" ")
		line.WriteString(styles.Warning.Render("NOT INSTALLED"))
	}

	// Show issues if any
	if len(status.Issues) > 0 {
		line.WriteString("  ")
		line.WriteString(styles.Warning.Render(status.Issues[0]))
	}

	return line.String()
}

// generateColorGradient generates a color gradient string based on color mode.
func generateColorGradient(colorMode string) string {
	// Generate ANSI color gradient based on terminal capability
	switch colorMode {
	case "truecolor":
		// True color gradient (24-bit)
		var gradient strings.Builder
		for i := 0; i < 16; i++ {
			r := i * 16
			g := 100 + i*8
			b := 200 - i*10
			gradient.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm█\x1b[0m", r, g, b))
		}
		return gradient.String()
	case "256":
		// 256-color gradient
		var gradient strings.Builder
		for i := 16; i < 32; i++ {
			gradient.WriteString(fmt.Sprintf("\x1b[38;5;%dm█\x1b[0m", i))
		}
		return gradient.String()
	default:
		// ANSI 16-color fallback
		return "\x1b[31m█\x1b[32m█\x1b[33m█\x1b[34m█\x1b[35m█\x1b[36m█\x1b[0m"
	}
}
