// Package tui provides tests for the TUI model.
package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/savanhi/shell/internal/detector"
	"github.com/savanhi/shell/internal/persistence"
)

func TestNewModel(t *testing.T) {
	m := NewModel()

	if m.CurrentScreen != ScreenWelcome {
		t.Errorf("expected initial screen to be ScreenWelcome, got %v", m.CurrentScreen)
	}

	if m.Ready {
		t.Error("expected Ready to be false initially")
	}

	if m.Quitting {
		t.Error("expected Quitting to be false initially")
	}

	if m.Selected == nil {
		t.Error("expected Selected map to be initialized")
	}

	if m.Loading {
		t.Error("expected Loading to be false initially")
	}

	if m.Error != nil {
		t.Error("expected Error to be nil initially")
	}
}

func TestModelInit(t *testing.T) {
	m := NewModel()
	cmd := m.Init()

	if cmd != nil {
		t.Error("expected Init() to return nil command")
	}
}

func TestScreenString(t *testing.T) {
	tests := []struct {
		screen   Screen
		expected string
	}{
		{ScreenWelcome, "Welcome"},
		{ScreenDetect, "Detect"},
		{ScreenThemeSelect, "ThemeSelect"},
		{ScreenFontSelect, "FontSelect"},
		{ScreenPreview, "Preview"},
		{ScreenInstall, "Install"},
		{ScreenComplete, "Complete"},
		{ScreenError, "Error"},
		{Screen(100), "Screen(100)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.screen.String(); got != tt.expected {
				t.Errorf("Screen(%d).String() = %q, want %q", tt.screen, got, tt.expected)
			}
		})
	}
}

func TestWithDetector(t *testing.T) {
	m := NewModel()
	result := &detector.DetectorResult{
		OS: &detector.OSInfo{
			Type:    detector.OSTypeLinux,
			Distro:  "ubuntu",
			Version: "22.04",
			Arch:    "amd64",
		},
		Shell: &detector.ShellInfo{
			Name:      detector.ShellTypeZsh,
			Version:   "5.9",
			IsDefault: true,
		},
	}

	m = m.WithDetector(result)

	if m.DetectorResult != result {
		t.Error("WithDetector did not set DetectorResult")
	}

	if m.systemInfo == nil {
		t.Error("WithDetector did not create systemInfo")
	}

	if m.systemInfo.OS == "" {
		t.Error("systemInfo.OS should not be empty")
	}
}

func TestWithDetectorNil(t *testing.T) {
	m := NewModel()
	m = m.WithDetector(nil)

	if m.DetectorResult != nil {
		t.Error("expected DetectorResult to be nil")
	}

	if m.systemInfo == nil {
		t.Error("expected systemInfo to be initialized even with nil result")
	}
}

func TestWithPersister(t *testing.T) {
	m := NewModel()
	// Create a mock persister
	var p persistence.Persister // nil is acceptable for this test

	m = m.WithPersister(p)

	if m.Persister != p {
		t.Error("WithPersister did not set Persister")
	}
}

func TestWithPreferences(t *testing.T) {
	m := NewModel()
	prefs := &persistence.Preferences{
		Theme: persistence.ThemePreferences{
			Name: "dark",
		},
	}

	m = m.WithPreferences(prefs)

	if m.Preferences != prefs {
		t.Error("WithPreferences did not set Preferences")
	}
}

func TestFormatSystemInfo(t *testing.T) {
	tests := []struct {
		name          string
		result        *detector.DetectorResult
		checkOS       string
		checkShell    string
		checkTerminal string
		checkFonts    int
	}{
		{
			name: "full detection result",
			result: &detector.DetectorResult{
				OS: &detector.OSInfo{
					Type:    detector.OSTypeMacOS,
					Distro:  "macOS",
					Version: "14.0",
					Arch:    "arm64",
				},
				Shell: &detector.ShellInfo{
					Name:    detector.ShellTypeZsh,
					Version: "5.9",
				},
				Terminal: &detector.TerminalInfo{
					Name:              "iTerm2",
					Version:           "3.4",
					SupportsTrueColor: true,
				},
				Fonts: &detector.FontInventory{
					NerdFonts: []detector.FontInfo{
						{Name: "FiraCode Nerd Font", IsNerdFont: true},
					},
				},
			},
			checkOS:       "macOS",
			checkShell:    "zsh",
			checkTerminal: "iTerm2",
			checkFonts:    1,
		},
		{
			name: "empty OS fields",
			result: &detector.DetectorResult{
				OS: &detector.OSInfo{
					Type: detector.OSTypeLinux,
					// Distro and Version are empty
				},
			},
			checkOS: "linux",
		},
		{
			name:   "nil result",
			result: nil,
		},
		{
			name: "partial result with terminal version",
			result: &detector.DetectorResult{
				Terminal: &detector.TerminalInfo{
					Name:    "kitty",
					Version: "0.30",
				},
			},
			checkTerminal: "kitty 0.30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := formatSystemInfo(tt.result)

			if tt.result == nil {
				if info.OS != "" || info.Shell != "" || info.Terminal != "" {
					t.Error("expected empty SystemInfo for nil result")
				}
				return
			}

			if tt.checkOS != "" && info.OS == "" {
				t.Error("expected OS to be set")
			}

			if tt.checkShell != "" && info.Shell == "" {
				t.Error("expected Shell to be set")
			}

			if tt.checkTerminal != "" && info.Terminal == "" {
				t.Error("expected Terminal to be set")
			}

			if tt.checkFonts > 0 && len(info.Fonts) != tt.checkFonts {
				t.Errorf("expected %d fonts, got %d", tt.checkFonts, len(info.Fonts))
			}
		})
	}
}

func TestModelUpdateWindowSize(t *testing.T) {
	m := NewModel()
	m.Width = 0
	m.Height = 0

	newModel, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = newModel.(Model)

	if m.Width != 80 {
		t.Errorf("expected Width to be 80, got %d", m.Width)
	}

	if m.Height != 24 {
		t.Errorf("expected Height to be 24, got %d", m.Height)
	}

	if !m.Ready {
		t.Error("expected Ready to be true after WindowSizeMsg")
	}

	if cmd != nil {
		t.Error("expected no command from WindowSizeMsg")
	}
}

func TestModelUpdateSystemDetectedMsg(t *testing.T) {
	m := NewModel()
	m.Loading = true
	m.LoadingMessage = "Detecting..."

	result := &detector.DetectorResult{
		OS: &detector.OSInfo{
			Type: detector.OSTypeLinux,
		},
	}

	newModel, cmd := m.Update(SystemDetectedMsg{Result: result})
	m = newModel.(Model)

	if m.DetectorResult != result {
		t.Error("expected DetectorResult to be set")
	}

	if m.Loading {
		t.Error("expected Loading to be false after SystemDetectedMsg")
	}

	if m.systemInfo == nil {
		t.Error("expected systemInfo to be set")
	}

	if cmd != nil {
		t.Error("expected no command from SystemDetectedMsg")
	}
}

func TestModelUpdateErrorMsg(t *testing.T) {
	m := NewModel()
	testErr := errors.New("test error")

	newModel, cmd := m.Update(ErrorMsg{Err: testErr})
	m = newModel.(Model)

	if m.Error != testErr {
		t.Error("expected Error to be set")
	}

	if m.CurrentScreen != ScreenError {
		t.Errorf("expected CurrentScreen to be ScreenError, got %v", m.CurrentScreen)
	}

	if cmd != nil {
		t.Error("expected no command from ErrorMsg")
	}
}

func TestModelUpdateLoadingMsg(t *testing.T) {
	m := NewModel()

	newModel, cmd := m.Update(LoadingMsg{Message: "Loading data..."})
	m = newModel.(Model)

	if !m.Loading {
		t.Error("expected Loading to be true")
	}

	if m.LoadingMessage != "Loading data..." {
		t.Errorf("expected LoadingMessage to be 'Loading data...', got %q", m.LoadingMessage)
	}

	if cmd != nil {
		t.Error("expected no command from LoadingMsg")
	}
}

// TestMessages tests that custom message types are properly defined
func TestMessageTypes(t *testing.T) {
	result := &detector.DetectorResult{}
	sysMsg := SystemDetectedMsg{Result: result}
	if sysMsg.Result != result {
		t.Error("SystemDetectedMsg.Result not set correctly")
	}

	testErr := errors.New("test")
	errMsg := ErrorMsg{Err: testErr}
	if errMsg.Err != testErr {
		t.Error("ErrorMsg.Err not set correctly")
	}

	loadMsg := LoadingMsg{Message: "test"}
	if loadMsg.Message != "test" {
		t.Error("LoadingMsg.Message not set correctly")
	}
}
