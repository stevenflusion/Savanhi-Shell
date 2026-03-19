// Package tui provides the Bubble Tea TUI for Savanhi Shell.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/savanhi/shell/internal/detector"
	"github.com/savanhi/shell/internal/persistence"
)

// Screen represents a screen in the TUI.
type Screen int

const (
	ScreenWelcome Screen = iota
	ScreenDetect
	ScreenThemeSelect
	ScreenFontSelect
	ScreenPreview
	ScreenInstall
	ScreenComplete
	ScreenHealthDashboard
	ScreenError
)

func (s Screen) String() string {
	switch s {
	case ScreenWelcome:
		return "Welcome"
	case ScreenDetect:
		return "Detect"
	case ScreenThemeSelect:
		return "ThemeSelect"
	case ScreenFontSelect:
		return "FontSelect"
	case ScreenPreview:
		return "Preview"
	case ScreenInstall:
		return "Install"
	case ScreenComplete:
		return "Complete"
	case ScreenHealthDashboard:
		return "HealthDashboard"
	case ScreenError:
		return "Error"
	default:
		return fmt.Sprintf("Screen(%d)", int(s))
	}
}

// Model is the main Bubble Tea model.
type Model struct {
	// Current screen
	CurrentScreen Screen

	// System detection results
	DetectorResult *detector.DetectorResult

	// Persistence layer
	Persister persistence.Persister

	// User preferences
	Preferences *persistence.Preferences

	// Health dashboard data
	HealthData *HealthData

	// Detected system information (for display)
	systemInfo *SystemInfo

	// Viewport dimensions
	Width  int
	Height int

	// Cursor position for lists
	Cursor int

	// Available items for selection (themes or fonts depending on screen)
	Items []string

	// Selected items (for multi-select)
	Selected map[string]bool

	// Theme selection
	SelectedTheme string

	// Font selection
	SelectedFont string

	// Loading state
	Loading        bool
	LoadingMessage string

	// Error state
	Error error

	// Ready flag (set after initial setup)
	Ready bool

	// Quit flag
	Quitting bool
}

// SystemInfo contains formatted system information for display.
type SystemInfo struct {
	OS       string
	Shell    string
	Terminal string
	Fonts    []string
	Tools    []string
}

// NewModel creates a new TUI model.
func NewModel() Model {
	return Model{
		CurrentScreen: ScreenWelcome,
		Selected:      make(map[string]bool),
		Ready:         false,
		Quitting:      false,
	}
}

// WithDetector sets the detector result.
func (m Model) WithDetector(result *detector.DetectorResult) Model {
	m.DetectorResult = result
	m.systemInfo = formatSystemInfo(result)
	return m
}

// WithPersister sets the persistence layer.
func (m Model) WithPersister(p persistence.Persister) Model {
	m.Persister = p
	return m
}

// WithPreferences sets the user preferences.
func (m Model) WithPreferences(prefs *persistence.Preferences) Model {
	m.Preferences = prefs
	return m
}

// WithThemes sets the available themes for selection.
func (m Model) WithThemes(themes []string) Model {
	m.Items = themes
	return m
}

// WithFonts sets the available fonts for selection.
func (m Model) WithFonts(fonts []string) Model {
	// Store fonts separately - Items is used for current screen
	// This will be used when transitioning to font selection screen
	return m
}

// Init implements tea.Model.Init.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.Ready = true
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case SystemDetectedMsg:
		m.DetectorResult = msg.Result
		m.systemInfo = formatSystemInfo(msg.Result)
		m.Loading = false
		return m, nil

	case ErrorMsg:
		m.Error = msg.Err
		m.CurrentScreen = ScreenError
		return m, nil

	case LoadingMsg:
		m.Loading = true
		m.LoadingMessage = msg.Message
		return m, nil

	case HealthCheckCompleteMsg:
		m.HealthData = msg.Data
		m.Loading = false
		if msg.Err != nil {
			m.Error = msg.Err
			m.CurrentScreen = ScreenError
		}
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model.View.
func (m Model) View() string {
	if m.Loading {
		return m.renderLoading()
	}

	switch m.CurrentScreen {
	case ScreenWelcome:
		return m.renderWelcome()
	case ScreenDetect:
		return m.renderDetect()
	case ScreenThemeSelect:
		return m.renderThemeSelect()
	case ScreenFontSelect:
		return m.renderFontSelect()
	case ScreenPreview:
		return m.renderPreview()
	case ScreenInstall:
		return m.renderInstall()
	case ScreenComplete:
		return m.renderComplete()
	case ScreenHealthDashboard:
		return m.renderHealthDashboard()
	case ScreenError:
		return m.renderError()
	default:
		return m.renderWelcome()
	}
}

// formatSystemInfo formats detector results for display.
func formatSystemInfo(result *detector.DetectorResult) *SystemInfo {
	if result == nil {
		return &SystemInfo{}
	}

	info := &SystemInfo{}

	if result.OS != nil {
		info.OS = fmt.Sprintf("%s %s (%s)", result.OS.Distro, result.OS.Version, result.OS.Arch)
		if info.OS == " ()" {
			info.OS = string(result.OS.Type)
		}
	}

	if result.Shell != nil {
		info.Shell = fmt.Sprintf("%s %s", result.Shell.Name, result.Shell.Version)
	}

	if result.Terminal != nil {
		info.Terminal = result.Terminal.Name
		if result.Terminal.Version != "" {
			info.Terminal += " " + result.Terminal.Version
		}
	}

	if result.Fonts != nil {
		for _, font := range result.Fonts.NerdFonts {
			info.Fonts = append(info.Fonts, font.Name)
		}
	}

	return info
}

// Messages for Bubble Tea.

// SystemDetectedMsg is sent when system detection completes.
type SystemDetectedMsg struct {
	Result *detector.DetectorResult
}

// ErrorMsg is sent when an error occurs.
type ErrorMsg struct {
	Err error
}

// LoadingMsg is sent when loading starts.
type LoadingMsg struct {
	Message string
}
