// Package views provides TUI views for Savanhi Shell.
// This file implements the install screen views.
package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/savanhi/shell/internal/installer"
)

// InstallModel represents the install screen state.
type InstallModel struct {
	// flow is the installation flow.
	flow *installer.InstallationFlow

	// components to install.
	components []string

	// progress tracks installation progress.
	progress *installer.InstallationProgress

	// phase is the current installation phase.
	phase string

	// step is the current step description.
	step string

	// percent is the completion percentage.
	percent float64

	// status is the overall status (running, success, failed).
	status string

	// errors are any errors encountered.
	errors []string

	// warnings are any warnings encountered.
	warnings []string

	// viewMode is the current view mode (progress, success, error).
	viewMode string

	// shellType is the detected shell type (bash, zsh, fish).
	shellType string

	// styles are the UI styles.
	styles InstallStyles

	// width is the terminal width.
	width int

	// height is the terminal height.
	height int
}

// InstallView is a view for installation screens.
type InstallView struct {
	// model is the install model.
	model InstallModel

	// styles are the UI styles.
	styles InstallStyles
}

// InstallStyles contains styles for the install screen.
type InstallStyles struct {
	// Title is the title style.
	Title lipgloss.Style

	// Phase is the phase style.
	Phase lipgloss.Style

	// Step is the step style.
	Step lipgloss.Style

	// Progress is the progress bar style.
	Progress lipgloss.Style

	// ProgressBg is the progress bar background style.
	ProgressBg lipgloss.Style

	// Success is the success message style.
	Success lipgloss.Style

	// Error is the error message style.
	Error lipgloss.Style

	// Warning is the warning message style.
	Warning lipgloss.Style

	// Component is the component status style.
	Component lipgloss.Style

	// Help is the help text style.
	Help lipgloss.Style
}

// NewInstallModel creates a new install model.
func NewInstallModel(flow *installer.InstallationFlow, components []string) InstallModel {
	view := NewInstallView()
	return InstallModel{
		flow:       flow,
		components: components,
		progress:   flow.GetProgress(),
		status:     "running",
		viewMode:   "progress",
		styles:     view.styles,
	}
}

// NewInstallView creates a new install view.
func NewInstallView() *InstallView {
	return &InstallView{
		styles: InstallStyles{
			Title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("62")).
				Padding(0, 1).
				MarginBottom(1),
			Phase: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("81")),
			Step: lipgloss.NewStyle().
				Foreground(lipgloss.Color("254")),
			Progress: lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Background(lipgloss.Color("238")),
			ProgressBg: lipgloss.NewStyle().
				Background(lipgloss.Color("238")),
			Success: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("42")),
			Error: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196")),
			Warning: lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")),
			Component: lipgloss.NewStyle().
				Foreground(lipgloss.Color("254")),
			Help: lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")),
		},
	}
}

// Init initializes the model.
func (m InstallModel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case ProgressUpdate:
		m.progress = msg.Progress
		m.phase = msg.Progress.CurrentPhase
		m.step = msg.Progress.CurrentStep
		m.percent = msg.Progress.Percent
		m.errors = msg.Progress.Errors
		m.warnings = msg.Progress.Warnings

		if m.progress.CurrentPhase == "completed" {
			m.status = "success"
			m.viewMode = "success"
		}
		return m, nil

	case InstallationComplete:
		m.status = "success"
		m.viewMode = "success"
		return m, nil

	case InstallationFailed:
		m.status = "failed"
		m.viewMode = "error"
		m.errors = msg.Errors
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.viewMode == "success" || m.viewMode == "error" {
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

// View renders the install screen.
func (m InstallModel) View() string {
	var b strings.Builder

	switch m.viewMode {
	case "progress":
		b.WriteString(m.renderProgressView())
	case "success":
		b.WriteString(m.renderSuccessView())
	case "error":
		b.WriteString(m.renderErrorView())
	default:
		b.WriteString(m.renderProgressView())
	}

	return b.String()
}

// renderProgressView renders the progress view.
func (m InstallModel) renderProgressView() string {
	var b strings.Builder

	// Title
	title := "📦 Installing Savanhi Shell"
	if m.phase != "" {
		title = fmt.Sprintf("📦 %s", strings.Title(m.phase))
	}
	b.WriteString(m.styles.Title.Render(title))
	b.WriteString("\n\n")

	// Current step
	if m.step != "" {
		stepText := fmt.Sprintf("▶ %s", m.step)
		b.WriteString(m.styles.Step.Render(stepText))
		b.WriteString("\n\n")
	}

	// Progress bar
	barWidth := 40
	filled := int(float64(barWidth) * m.percent / 100)
	empty := barWidth - filled

	progressBar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	b.WriteString(m.styles.ProgressBg.Render(progressBar))
	b.WriteString(fmt.Sprintf(" %.0f%%", m.percent))
	b.WriteString("\n\n")

	// Components
	b.WriteString("Components:\n")
	for _, comp := range m.components {
		icon := "○"
		if m.progress != nil {
			// Check if this component is being installed
			if strings.Contains(strings.ToLower(m.step), strings.ToLower(comp)) {
				icon = "◐"
			}
		}
		b.WriteString(fmt.Sprintf("  %s %s\n", m.styles.Component.Render(icon), comp))
	}
	b.WriteString("\n")

	// Warnings
	if len(m.warnings) > 0 {
		b.WriteString(m.styles.Warning.Render("⚠ Warnings:\n"))
		for _, w := range m.warnings {
			b.WriteString(fmt.Sprintf("  • %s\n", w))
		}
		b.WriteString("\n")
	}

	// Help
	b.WriteString(m.styles.Help.Render("\nPress Ctrl+C to cancel"))

	return b.String()
}

// renderSuccessView renders the success view.
func (m InstallModel) renderSuccessView() string {
	var b strings.Builder

	// Title
	b.WriteString(m.styles.Title.Render("✅ Installation Complete"))
	b.WriteString("\n\n")

	// Success message
	b.WriteString(m.styles.Success.Render("Savanhi Shell has been successfully installed!"))
	b.WriteString("\n\n")

	// Installed components
	b.WriteString("Installed components:\n")
	for _, comp := range m.components {
		b.WriteString(fmt.Sprintf("  ✓ %s\n", comp))
	}
	b.WriteString("\n")

	// Next steps
	b.WriteString(m.styles.Step.Render("Next steps:\n"))

	// Get the appropriate RC source command based on shell type
	rcSource := getRCPathMessage(m.shellType)
	b.WriteString(fmt.Sprintf("  1. Restart your shell or run: %s\n", rcSource))
	b.WriteString("  2. Choose a theme: savanhi theme list\n")
	b.WriteString("  3. Configure your preferences: savanhi config\n")
	b.WriteString("\n")

	// Warnings from installation
	if len(m.warnings) > 0 {
		b.WriteString(m.styles.Warning.Render("Note:\n"))
		for _, w := range m.warnings {
			b.WriteString(fmt.Sprintf("  • %s\n", w))
		}
		b.WriteString("\n")
	}

	// Help
	b.WriteString(m.styles.Help.Render("Press Enter to exit"))

	return b.String()
}

// renderErrorView renders the error view.
func (m InstallModel) renderErrorView() string {
	var b strings.Builder

	// Title
	b.WriteString(m.styles.Title.Render("❌ Installation Failed"))
	b.WriteString("\n\n")

	// Error message
	b.WriteString(m.styles.Error.Render("The installation encountered errors and could not complete."))
	b.WriteString("\n\n")

	// Errors
	if len(m.errors) > 0 {
		b.WriteString(m.styles.Error.Render("Errors:\n"))
		for _, e := range m.errors {
			b.WriteString(fmt.Sprintf("  • %s\n", e))
		}
		b.WriteString("\n")
	}

	// Rollback info
	b.WriteString(m.styles.Step.Render("Your system has been rolled back to its previous state.\n"))
	b.WriteString("\n")

	// Help text
	b.WriteString(m.styles.Help.Render("Press Enter to exit or 'r' to retry"))

	return b.String()
}

// ProgressUpdate is a message for progress updates.
type ProgressUpdate struct {
	Progress *installer.InstallationProgress
}

// InstallationComplete is a message for installation completion.
type InstallationComplete struct{}

// InstallationFailed is a message for installation failure.
type InstallationFailed struct {
	Errors []string
}

// SetProgress updates the progress.
func (m *InstallModel) SetProgress(progress *installer.InstallationProgress) {
	m.progress = progress
	m.phase = progress.CurrentPhase
	m.step = progress.CurrentStep
	m.percent = progress.Percent
	m.errors = progress.Errors
	m.warnings = progress.Warnings
}

// SetStatus sets the installation status.
func (m *InstallModel) SetStatus(status string) {
	m.status = status
	if status == "success" {
		m.viewMode = "success"
	} else if status == "failed" {
		m.viewMode = "error"
	}
}

// SetViewMode sets the view mode.
func (m *InstallModel) SetViewMode(mode string) {
	m.viewMode = mode
}

// GetProgress returns the current progress.
func (m InstallModel) GetProgress() *installer.InstallationProgress {
	return m.progress
}

// GetStatus returns the installation status.
func (m InstallModel) GetStatus() string {
	return m.status
}

// GetErrors returns the errors list.
func (m InstallModel) GetErrors() []string {
	return m.errors
}

// GetWarnings returns the warnings list.
func (m InstallModel) GetWarnings() []string {
	return m.warnings
}

// IsComplete returns whether the installation is complete.
func (m InstallModel) IsComplete() bool {
	return m.status == "success" || m.status == "failed"
}

// IsSuccess returns whether the installation succeeded.
func (m InstallModel) IsSuccess() bool {
	return m.status == "success"
}

// IsFailed returns whether the installation failed.
func (m InstallModel) IsFailed() bool {
	return m.status == "failed"
}

// View renders the install view.
func (v *InstallView) View(m InstallModel) string {
	return m.View()
}

// GetStyles returns the styles.
func (v *InstallView) GetStyles() InstallStyles {
	return v.styles
}

// SetStyles sets the styles.
func (v *InstallView) SetStyles(styles InstallStyles) {
	v.styles = styles
}

// SetShellType sets the shell type for the install model.
func (m *InstallModel) SetShellType(shellType string) {
	m.shellType = shellType
}

// getRCPathMessage returns the appropriate RC file source command based on shell type.
func getRCPathMessage(shellType string) string {
	switch shellType {
	case "fish":
		return "source ~/.config/fish/config.fish"
	case "bash":
		return "source ~/.bashrc"
	case "zsh":
		return "source ~/.zshrc"
	default:
		return "source ~/.zshrc (or ~/.bashrc)"
	}
}
