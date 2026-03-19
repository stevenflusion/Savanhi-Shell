// Package tui provides update logic for the TUI.
package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/savanhi/shell/internal/detector"
	"github.com/savanhi/shell/internal/preview"
)

// handleKeyPress handles key press events.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle global keys
	if IsQuitKey(msg) {
		m.Quitting = true
		return m, tea.Quit
	}

	// Handle screen-specific keys
	switch m.CurrentScreen {
	case ScreenWelcome:
		return m.handleWelcomeKeys(msg)
	case ScreenDetect:
		return m.handleDetectKeys(msg)
	case ScreenThemeSelect:
		return m.handleThemeSelectKeys(msg)
	case ScreenFontSelect:
		return m.handleFontSelectKeys(msg)
	case ScreenPreview:
		return m.handlePreviewKeys(msg)
	case ScreenInstall:
		return m.handleInstallKeys(msg)
	case ScreenComplete:
		return m.handleCompleteKeys(msg)
	case ScreenHealthDashboard:
		return m.handleHealthDashboardKeys(msg)
	case ScreenError:
		return m.handleErrorKeys(msg)
	default:
		return m, nil
	}
}

// handleWelcomeKeys handles key presses on the welcome screen.
func (m Model) handleWelcomeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if IsSelectionKey(msg) || IsConfirmKey(msg) {
		m.CurrentScreen = ScreenDetect
		m.Loading = true
		m.LoadingMessage = "Detecting system information..."
		return m, detectSystem()
	}
	return m, nil
}

// handleDetectKeys handles key presses on the detection screen.
func (m Model) handleDetectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if IsSelectionKey(msg) || IsConfirmKey(msg) {
		// Move to theme selection after confirming detection
		m.CurrentScreen = ScreenThemeSelect
		return m, nil
	}

	// 'h' or 'H' to go to Health Dashboard
	switch msg.String() {
	case "h", "H":
		m.CurrentScreen = ScreenHealthDashboard
		m.Loading = true
		m.LoadingMessage = "Running health checks..."
		return m, RunHealthCheck()
	}

	return m, nil
}

// handleThemeSelectKeys handles key presses on the theme selection screen.
func (m Model) handleThemeSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, Keys.Up) {
		if m.Cursor > 0 {
			m.Cursor--
		}
		return m, nil
	}

	if key.Matches(msg, Keys.Down) {
		if m.Cursor < len(m.Items)-1 {
			m.Cursor++
		}
		return m, nil
	}

	if IsSelectionKey(msg) || IsConfirmKey(msg) {
		if len(m.Items) > 0 && m.Cursor < len(m.Items) {
			// Store selected theme
			m.Selected["theme"] = true // Mark theme as selected
			m.SelectedTheme = m.Items[m.Cursor]
			// Load fonts for font selection screen
			m.Items = preview.GetRecommendedFonts()
			m.CurrentScreen = ScreenFontSelect
			m.Cursor = 0
		}
		return m, nil
	}

	if IsCancelKey(msg) {
		m.CurrentScreen = ScreenDetect
		m.Cursor = 0
		return m, nil
	}

	return m, nil
}

// handleFontSelectKeys handles key presses on the font selection screen.
func (m Model) handleFontSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, Keys.Up) {
		if m.Cursor > 0 {
			m.Cursor--
		}
		return m, nil
	}

	if key.Matches(msg, Keys.Down) {
		if m.Cursor < len(m.Items)-1 {
			m.Cursor++
		}
		return m, nil
	}

	if IsSelectionKey(msg) || IsConfirmKey(msg) {
		if len(m.Items) > 0 && m.Cursor < len(m.Items) {
			// Store selected font
			m.Selected["font"] = true // Mark font as selected
			m.SelectedFont = m.Items[m.Cursor]
			m.CurrentScreen = ScreenPreview
			m.Cursor = 0
		}
		return m, nil
	}

	if IsCancelKey(msg) {
		m.CurrentScreen = ScreenThemeSelect
		m.Cursor = 0
		// Restore themes list
		m.Items = preview.GetBundledThemes()
		return m, nil
	}

	return m, nil
}

// handlePreviewKeys handles key presses on the preview screen.
func (m Model) handlePreviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if IsConfirmKey(msg) {
		m.CurrentScreen = ScreenInstall
		return m, nil
	}

	if IsCancelKey(msg) {
		m.CurrentScreen = ScreenFontSelect
		return m, nil
	}

	return m, nil
}

// handleInstallKeys handles key presses on the install screen.
func (m Model) handleInstallKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// During installation, only allow cancel
	if m.Loading {
		if IsCancelKey(msg) {
			m.CurrentScreen = ScreenPreview
			return m, nil
		}
		return m, nil
	}

	if IsConfirmKey(msg) {
		m.CurrentScreen = ScreenComplete
		return m, nil
	}

	if IsCancelKey(msg) {
		m.CurrentScreen = ScreenPreview
		return m, nil
	}

	return m, nil
}

// handleCompleteKeys handles key presses on the complete screen.
func (m Model) handleCompleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if IsQuitKey(msg) || IsConfirmKey(msg) {
		m.Quitting = true
		return m, tea.Quit
	}
	return m, nil
}

// handleErrorKeys handles key presses on the error screen.
func (m Model) handleErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if IsCancelKey(msg) {
		m.Error = nil
		m.CurrentScreen = ScreenWelcome
		return m, nil
	}

	if IsQuitKey(msg) {
		m.Quitting = true
		return m, tea.Quit
	}

	return m, nil
}

// handleHealthDashboardKeys handles key presses on the health dashboard screen.
// Supports: R (refresh), E (export), Q/Esc (quit/back).
func (m Model) handleHealthDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r", "R":
		// Refresh health check
		return m, RunHealthCheck()

	case "e", "E":
		// Export health report
		if m.HealthData != nil && m.HealthData.ExportPath != "" {
			err := ExportHealthReport(m.HealthData, m.HealthData.ExportPath)
			if err != nil {
				m.Error = err
				m.CurrentScreen = ScreenError
				return m, nil
			}
		}
		return m, nil

	case "q", "Q":
		// Quit
		m.Quitting = true
		return m, tea.Quit

	case "esc":
		// Go back to previous screen (Detect or Welcome)
		if m.DetectorResult != nil {
			m.CurrentScreen = ScreenDetect
		} else {
			m.CurrentScreen = ScreenWelcome
		}
		return m, nil

	case "?":
		// Show help (for now, just stay on screen)
		// TODO: Implement help modal
		return m, nil
	}

	return m, nil
}

// detectSystem is a command that detects system information.
func detectSystem() tea.Cmd {
	return func() tea.Msg {
		// In a real implementation, this would call the detector
		// For now, return a placeholder result
		d := detector.NewDefaultDetector()
		result, err := d.DetectAll()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return SystemDetectedMsg{Result: result}
	}
}
