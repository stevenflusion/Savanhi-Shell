// Package views provides TUI screen components.
package views

import (
	"fmt"
	"strings"

	"github.com/savanhi/shell/internal/detector"
	"github.com/savanhi/shell/internal/tui/styles"
)

// DetectionModel represents the state of the detection screen.
type DetectionModel struct {
	// Detection results from the detector
	Result *detector.DetectorResult

	// Whether detection is in progress
	Loading bool

	// Any error that occurred during detection
	Error error

	// Cursor position for navigation
	Cursor int

	// Available actions
	Actions []string
}

// NewDetectionModel creates a new detection model.
func NewDetectionModel() DetectionModel {
	return DetectionModel{
		Loading: true,
		Actions: []string{
			"Continue",
			"Refresh",
			"Back",
		},
		Cursor: 0,
	}
}

// WithResult sets the detection result.
func (m DetectionModel) WithResult(result *detector.DetectorResult) DetectionModel {
	m.Result = result
	m.Loading = false
	return m
}

// WithError sets the error.
func (m DetectionModel) WithError(err error) DetectionModel {
	m.Error = err
	m.Loading = false
	return m
}

// SetCursor sets the cursor position.
func (m DetectionModel) SetCursor(pos int) DetectionModel {
	if pos >= 0 && pos < len(m.Actions) {
		m.Cursor = pos
	}
	return m
}

// View renders the detection screen.
func (m DetectionModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString("\n")
	b.WriteString(styles.Title.Render(" System Detection "))
	b.WriteString("\n\n")

	if m.Loading {
		b.WriteString(m.renderLoading())
	} else if m.Error != nil {
		b.WriteString(m.renderError())
	} else {
		b.WriteString(m.renderResults())
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderLoading renders the loading state.
func (m DetectionModel) renderLoading() string {
	var b strings.Builder

	b.WriteString(styles.Box.Render(
		styles.Info.Render(" Detecting system information...") + "\n\n" +
			styles.Muted.Render(" This may take a few seconds."),
	))

	return b.String()
}

// renderError renders the error state.
func (m DetectionModel) renderError() string {
	var b strings.Builder

	b.WriteString(styles.Box.Render(
		styles.Error.Render(" Detection Failed") + "\n\n" +
			styles.Muted.Render(m.Error.Error()),
	))

	return b.String()
}

// renderResults renders the detection results.
func (m DetectionModel) renderResults() string {
	var b strings.Builder

	// Detection container
	containerStyle := styles.DetectionBox.Container

	var content strings.Builder

	// OS Information
	content.WriteString(m.renderSectionHeader("Operating System"))
	if m.Result != nil && m.Result.OS != nil {
		content.WriteString(m.renderOSInfo(m.Result.OS))
	} else {
		content.WriteString(styles.Warning.Render("  Not detected"))
	}
	content.WriteString("\n")

	// Shell Information
	content.WriteString(m.renderSectionHeader("Shell"))
	if m.Result != nil && m.Result.Shell != nil {
		content.WriteString(m.renderShellInfo(m.Result.Shell))
	} else {
		content.WriteString(styles.Warning.Render("  Not detected"))
	}
	content.WriteString("\n")

	// Terminal Information
	content.WriteString(m.renderSectionHeader("Terminal"))
	if m.Result != nil && m.Result.Terminal != nil {
		content.WriteString(m.renderTerminalInfo(m.Result.Terminal))
	} else {
		content.WriteString(styles.Warning.Render("  Not detected"))
	}
	content.WriteString("\n")

	// Font Information
	content.WriteString(m.renderSectionHeader("Fonts"))
	if m.Result != nil && m.Result.Fonts != nil {
		content.WriteString(m.renderFontInfo(m.Result.Fonts))
	} else {
		content.WriteString(styles.Warning.Render("  Not detected"))
	}
	content.WriteString("\n")

	// Existing Components
	if m.Result != nil && m.Result.ExistingConfigs != nil {
		content.WriteString(m.renderSectionHeader("Installed Components"))
		content.WriteString(m.renderExistingConfigs(m.Result.ExistingConfigs))
	}

	b.WriteString(containerStyle.Render(content.String()))
	b.WriteString("\n")

	// Actions
	b.WriteString(m.renderActions())

	return b.String()
}

// renderSectionHeader renders a section header.
func (m DetectionModel) renderSectionHeader(title string) string {
	return styles.Subtitle.Render(title) + "\n"
}

// renderOSInfo renders OS information.
func (m DetectionModel) renderOSInfo(info *detector.OSInfo) string {
	var b strings.Builder

	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("Type:"))
	b.WriteString(" ")
	b.WriteString(styles.DetectionBox.Value.Render(string(info.Type)))
	b.WriteString("\n")

	if info.Distro != "" {
		b.WriteString("  ")
		b.WriteString(styles.DetectionBox.Label.Render("Distro:"))
		b.WriteString(" ")
		b.WriteString(styles.DetectionBox.Value.Render(info.Distro))
		b.WriteString(" ")
		b.WriteString(styles.Muted.Render(info.Version))
		b.WriteString("\n")
	}

	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("Arch:"))
	b.WriteString(" ")
	b.WriteString(styles.DetectionBox.Value.Render(info.Arch))
	b.WriteString("\n")

	if info.PackageMgr != "" {
		b.WriteString("  ")
		b.WriteString(styles.DetectionBox.Label.Render("Package Mgr:"))
		b.WriteString(" ")
		b.WriteString(styles.DetectionBox.Value.Render(info.PackageMgr))
		b.WriteString("\n")
	}

	return b.String()
}

// renderShellInfo renders shell information.
func (m DetectionModel) renderShellInfo(info *detector.ShellInfo) string {
	var b strings.Builder

	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("Name:"))
	b.WriteString(" ")
	b.WriteString(styles.DetectionBox.Value.Render(string(info.Name)))
	b.WriteString("\n")

	if info.Version != "" {
		b.WriteString("  ")
		b.WriteString(styles.DetectionBox.Label.Render("Version:"))
		b.WriteString(" ")
		b.WriteString(styles.DetectionBox.Value.Render(info.Version))
		b.WriteString("\n")
	}

	if info.RCFile != "" {
		b.WriteString("  ")
		b.WriteString(styles.DetectionBox.Label.Render("RC File:"))
		b.WriteString(" ")
		b.WriteString(styles.Muted.Render(info.RCFile))
		b.WriteString("\n")
	}

	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("Default:"))
	b.WriteString(" ")
	if info.IsDefault {
		b.WriteString(styles.StatusStyle.Success.Render("✓ Yes"))
	} else {
		b.WriteString(styles.Muted.Render("No"))
	}
	b.WriteString("\n")

	return b.String()
}

// renderTerminalInfo renders terminal information.
func (m DetectionModel) renderTerminalInfo(info *detector.TerminalInfo) string {
	var b strings.Builder

	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("Name:"))
	b.WriteString(" ")
	b.WriteString(styles.DetectionBox.Value.Render(info.Name))
	b.WriteString("\n")

	if info.Version != "" {
		b.WriteString("  ")
		b.WriteString(styles.DetectionBox.Label.Render("Version:"))
		b.WriteString(" ")
		b.WriteString(styles.DetectionBox.Value.Render(info.Version))
		b.WriteString("\n")
	}

	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("True Color:"))
	b.WriteString(" ")
	if info.SupportsTrueColor {
		b.WriteString(styles.StatusStyle.Success.Render("✓"))
	} else {
		b.WriteString(styles.Muted.Render("✗"))
	}
	b.WriteString("\n")

	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("Ligatures:"))
	b.WriteString(" ")
	if info.SupportsLigatures {
		b.WriteString(styles.StatusStyle.Success.Render("✓"))
	} else {
		b.WriteString(styles.Muted.Render("✗"))
	}
	b.WriteString("\n")

	return b.String()
}

// renderFontInfo renders font information.
func (m DetectionModel) renderFontInfo(info *detector.FontInventory) string {
	var b strings.Builder

	// Nerd Fonts
	nerdCount := len(info.NerdFonts)
	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("Nerd Fonts:"))
	b.WriteString(" ")
	if nerdCount > 0 {
		b.WriteString(styles.StatusStyle.Success.Render(fmt.Sprintf("✓ %d found", nerdCount)))
		b.WriteString("\n")

		// List first few fonts
		maxList := 3
		count := 0
		for _, font := range info.NerdFonts {
			if count >= maxList {
				break
			}
			b.WriteString("    ")
			b.WriteString(styles.Muted.Render("- " + font.Name))
			b.WriteString("\n")
			count++
		}
		if nerdCount > maxList {
			b.WriteString("    ")
			b.WriteString(styles.Muted.Render(fmt.Sprintf("- ... and %d more", nerdCount-maxList)))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(styles.Warning.Render("None found"))
		b.WriteString("\n")
	}

	// Total fonts
	b.WriteString("  ")
	b.WriteString(styles.DetectionBox.Label.Render("Total Fonts:"))
	b.WriteString(" ")
	b.WriteString(styles.DetectionBox.Value.Render(fmt.Sprintf("%d", len(info.Fonts))))
	b.WriteString("\n")

	return b.String()
}

// renderExistingConfigs renders installed component information.
func (m DetectionModel) renderExistingConfigs(info *detector.ConfigSnapshot) string {
	var b strings.Builder

	components := []struct {
		name      string
		installed bool
	}{
		{"Oh My Posh", info.HasOhMyPosh},
		{"Starship", info.HasStarship},
	}

	for _, comp := range components {
		b.WriteString("  ")
		if comp.installed {
			b.WriteString(styles.StatusStyle.Success.Render("✓ " + comp.name))
		} else {
			b.WriteString(styles.Muted.Render("○ " + comp.name))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderActions renders available actions.
func (m DetectionModel) renderActions() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.Subtitle.Render("Actions:"))
	b.WriteString("\n")

	for i, action := range m.Actions {
		b.WriteString("  ")
		if i == m.Cursor {
			b.WriteString(styles.ItemStyle.Selected.Render("→ " + action))
		} else {
			b.WriteString(styles.ItemStyle.Normal.Render("  " + action))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderFooter renders the footer with keybindings.
func (m DetectionModel) renderFooter() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(styles.FooterStyle.Render(strings.Repeat("─", 50)))
	b.WriteString("\n")

	if m.Loading {
		b.WriteString(styles.Muted.Render("Press Esc to cancel"))
	} else if m.Error != nil {
		b.WriteString(styles.FormatKey("Esc", "Back"))
		b.WriteString("  ")
		b.WriteString(styles.FormatKey("q", "Quit"))
	} else {
		b.WriteString(styles.FormatKey("↑↓", "Navigate"))
		b.WriteString("  ")
		b.WriteString(styles.FormatKey("Enter", "Select"))
		b.WriteString("  ")
		b.WriteString(styles.FormatKey("q", "Quit"))
	}

	return b.String()
}

// GetAction returns the currently selected action.
func (m DetectionModel) GetAction() string {
	if m.Cursor >= 0 && m.Cursor < len(m.Actions) {
		return m.Actions[m.Cursor]
	}
	return ""
}

// DetectionAction represents an action on the detection screen.
type DetectionAction int

const (
	ActionContinue DetectionAction = iota
	ActionRefresh
	ActionBack
)

// DetermineAction maps action names to constants.
func DetermineAction(name string) DetectionAction {
	switch name {
	case "Continue":
		return ActionContinue
	case "Refresh":
		return ActionRefresh
	case "Back":
		return ActionBack
	default:
		return ActionContinue
	}
}
