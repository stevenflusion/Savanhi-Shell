// Package tui provides plugin selection screen rendering for the TUI.
package tui

import (
	"strings"

	"github.com/savanhi/shell/internal/installer"
	"github.com/savanhi/shell/internal/tui/styles"
)

// renderPluginSelect renders the plugin selection screen.
func (m Model) renderPluginSelect() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Select Zsh Plugins"))
	b.WriteString("\n\n")

	// Description
	b.WriteString(styles.Info.Render("Enhance your zsh experience with these recommended plugins."))
	b.WriteString("\n")
	b.WriteString(styles.Muted.Render("Press Space to toggle selection, Enter to continue."))
	b.WriteString("\n\n")

	if len(m.AvailablePlugins) == 0 {
		b.WriteString(styles.Warning.Render("No plugins detected."))
	} else {
		// Render each plugin as a checkbox item
		for i, status := range m.AvailablePlugins {
			b.WriteString(m.renderPluginItem(i, status))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.renderPluginFooter())

	return b.String()
}

// renderPluginItem renders a single plugin item with checkbox.
func (m Model) renderPluginItem(index int, status installer.PluginStatus) string {
	var b strings.Builder

	// Cursor indicator
	cursor := "  "
	if index == m.Cursor {
		cursor = "→ "
	}

	// Checkbox state
	checkbox := "[ ]"
	if m.SelectedPlugins[status.Plugin.Name] {
		checkbox = "[✓]"
	}

	// Build the line
	var line strings.Builder
	line.WriteString(cursor)
	line.WriteString(checkbox)
	line.WriteString(" ")

	// Plugin name
	name := status.Plugin.DisplayName
	if index == m.Cursor {
		line.WriteString(styles.ItemStyle.Selected.Render(name))
	} else {
		line.WriteString(styles.ItemStyle.Normal.Render(name))
	}

	// Status indicator
	if status.Installed {
		installedText := " (installed"
		if status.Method != installer.MethodNone {
			installedText += " via " + status.Method.String()
		}
		installedText += ")"
		b.WriteString(styles.StatusStyle.Success.Render(installedText))
	}

	b.WriteString(line.String())

	// Description on next line for selected item
	if index == m.Cursor {
		b.WriteString("\n")
		b.WriteString("    ")
		b.WriteString(styles.Muted.Render(status.Plugin.Description))
		if status.Plugin.MustBeLast {
			b.WriteString(" ")
			b.WriteString(styles.Warning.Render("[must be sourced last]"))
		}
	}

	return b.String()
}

// renderPluginFooter renders the footer with keybindings for plugin selection.
func (m Model) renderPluginFooter() string {
	var b strings.Builder

	b.WriteString(styles.FooterStyle.Render("─"))
	b.WriteString("\n")

	// Key hints
	keys := []string{
		styles.FormatKey("↑↓/jk", "Navigate"),
		styles.FormatKey("Space", "Toggle"),
		styles.FormatKey("Enter", "Continue"),
		styles.FormatKey("Esc", "Back"),
		styles.FormatKey("?", "Help"),
		styles.FormatKey("q", "Quit"),
	}

	b.WriteString(strings.Join(keys, "  "))
	b.WriteString("\n")

	// Show preview hint
	b.WriteString(styles.Muted.Render("Press Enter to preview changes before installation."))
	b.WriteString("\n")

	return b.String()
}

// renderPluginPreview renders a preview of what will be installed.
func (m Model) renderPluginPreview() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Title.Render("Plugin Installation Preview"))
	b.WriteString("\n\n")

	// Get selected plugins
	selectedPlugins := m.getSelectedPluginsList()

	if len(selectedPlugins) == 0 {
		b.WriteString(styles.Info.Render("No plugins selected for installation."))
		b.WriteString("\n")
	} else {
		b.WriteString(styles.Subtitle.Render("Selected Plugins:"))
		b.WriteString("\n\n")

		for _, status := range selectedPlugins {
			b.WriteString(m.renderPluginPreviewItem(status))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(styles.Subtitle.Render("Changes to .zshrc:"))
		b.WriteString("\n\n")

		b.WriteString(m.renderZshrcDiff())
	}

	b.WriteString("\n")
	b.WriteString(m.renderPreviewFooter())

	return b.String()
}

// renderPluginPreviewItem renders a single plugin preview item.
func (m Model) renderPluginPreviewItem(status installer.PluginStatus) string {
	var b strings.Builder

	// Plugin name with install method
	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Value.Render(status.Plugin.DisplayName))

	if status.Installed {
		b.WriteString(" ")
		b.WriteString(styles.StatusStyle.Success.Render("(already installed)"))
	} else {
		b.WriteString(" ")
		b.WriteString(styles.Muted.Render("→ will be installed via "))
		b.WriteString(styles.DetectionBox.Label.Render(status.Method.String()))
	}
	b.WriteString("\n")

	// Description
	b.WriteString("      ")
	b.WriteString(styles.Muted.Render(status.Plugin.Description))
	b.WriteString("\n")

	return b.String()
}

// renderZshrcDiff renders a preview of .zshrc changes.
func (m Model) renderZshrcDiff() string {
	var b strings.Builder

	selectedPlugins := m.getSelectedPluginsList()

	if len(selectedPlugins) == 0 {
		return styles.Muted.Render("  No changes to .zshrc")
	}

	// Check if any plugin needs to be installed (not already installed)
	hasNewPlugins := false
	for _, status := range selectedPlugins {
		if !status.Installed {
			hasNewPlugins = true
			break
		}
	}

	if !hasNewPlugins {
		return styles.Muted.Render("  All selected plugins are already installed. No changes needed.")
	}

	// Determine installation method (prefer OMZ if available)
	installMethod := installer.MethodGitClone
	for _, status := range selectedPlugins {
		if status.Method != installer.MethodNone && status.Method != installer.MethodGitClone {
			installMethod = status.Method
			break
		}
	}

	// Show preview based on installation method
	switch installMethod {
	case installer.MethodOhMyZsh:
		b.WriteString(styles.Muted.Render("  plugins=("))
		b.WriteString("\n")

		// Show existing plugins (placeholder)
		b.WriteString(styles.Muted.Render("    git\n"))
		b.WriteString(styles.Muted.Render("    ...\n"))

		// Show new plugins to add
		for _, status := range selectedPlugins {
			if !status.Installed {
				b.WriteString("    ")
				b.WriteString(styles.StatusStyle.Success.Render(status.Plugin.OhMyZshName))
				b.WriteString("\n")
			}
		}

		b.WriteString(styles.Muted.Render("  )"))

	case installer.MethodHomebrew, installer.MethodGitClone:
		// Show source lines
		for _, status := range selectedPlugins {
			if !status.Installed {
				b.WriteString(styles.Info.Render("  # Savanhi: "))
				b.WriteString(styles.DetectionBox.Label.Render(status.Plugin.DisplayName))
				b.WriteString("\n")
				b.WriteString(styles.Muted.Render("  source ~/.zsh/"))
				b.WriteString(styles.DetectionBox.Value.Render(status.Plugin.Name))
				b.WriteString(styles.Muted.Render("/" + status.Plugin.SourceFile))
				b.WriteString("\n")
				if status.Plugin.MustBeLast {
					b.WriteString(styles.Warning.Render("  # Note: This plugin must be sourced last"))
					b.WriteString("\n")
				}
			}
		}
	}

	return b.String()
}

// getSelectedPluginsList returns the list of selected plugin statuses.
func (m Model) getSelectedPluginsList() []installer.PluginStatus {
	var selected []installer.PluginStatus

	for _, status := range m.AvailablePlugins {
		if m.SelectedPlugins[status.Plugin.Name] {
			selected = append(selected, status)
		}
	}

	// Sort: plugins that must be last go to the end
	result := make([]installer.PluginStatus, 0, len(selected))
	mustBeLast := make([]installer.PluginStatus, 0)

	for _, status := range selected {
		if status.Plugin.MustBeLast {
			mustBeLast = append(mustBeLast, status)
		} else {
			result = append(result, status)
		}
	}

	// Append must-be-last plugins at the end
	result = append(result, mustBeLast...)

	return result
}

// renderPreviewFooter renders the footer for the preview screen.
func (m Model) renderPreviewFooter() string {
	var b strings.Builder

	b.WriteString(styles.FooterStyle.Render("─"))
	b.WriteString("\n")

	keys := []string{
		styles.FormatKey("Enter", "Install"),
		styles.FormatKey("Esc", "Back"),
		styles.FormatKey("q", "Quit"),
	}

	b.WriteString(strings.Join(keys, "  "))
	b.WriteString("\n")

	return b.String()
}

// WithPlugins sets the available plugins for the model.
func (m Model) WithPlugins(plugins []installer.PluginStatus) Model {
	m.AvailablePlugins = plugins

	// Pre-select plugins that are already installed
	for _, status := range plugins {
		// By default, select all recommended plugins
		// Users can deselect if they want
		m.SelectedPlugins[status.Plugin.Name] = true
	}

	return m
}
