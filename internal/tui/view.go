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

	// Preview placeholder
	b.WriteString(styles.Box.Render(
		styles.Info.Render("Preview will show your selected configuration") + "\n\n" +
			styles.Muted.Render("Press Enter to install or Esc to go back"),
	))

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
