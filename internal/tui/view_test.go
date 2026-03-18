// Package tui provides tests for TUI view rendering.
package tui

import (
	"strings"
	"testing"

	"github.com/savanhi/shell/internal/detector"
)

func TestRenderLoading(t *testing.T) {
	tests := []struct {
		name         string
		model        Model
		wantContains []string
	}{
		{
			name: "with loading message",
			model: Model{
				LoadingMessage: "Detecting system...",
			},
			wantContains: []string{"Savanhi Shell", "Detecting system..."},
		},
		{
			name: "without loading message",
			model: Model{
				LoadingMessage: "",
			},
			wantContains: []string{"Savanhi Shell", "Loading..."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.model.Loading = true
			output := tt.model.renderLoading()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("renderLoading() missing %q in output", want)
				}
			}
		})
	}
}

func TestRenderWelcome(t *testing.T) {
	m := NewModel()
	output := m.renderWelcome()

	wantContains := []string{
		"Savanhi Shell",
		"Welcome",
		"Press Enter to begin",
	}

	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("renderWelcome() missing %q in output", want)
		}
	}
}

func TestRenderDetect(t *testing.T) {
	tests := []struct {
		name            string
		model           Model
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "with detection results",
			model: Model{
				CurrentScreen: ScreenDetect,
				systemInfo: &SystemInfo{
					OS:       "Ubuntu 22.04 (amd64)",
					Shell:    "zsh 5.9",
					Terminal: "kitty 0.30",
					Fonts:    []string{"FiraCode Nerd Font"},
				},
				DetectorResult: &detector.DetectorResult{
					OS: &detector.OSInfo{
						Type:    detector.OSTypeLinux,
						Distro:  "Ubuntu",
						Version: "22.04",
						Arch:    "amd64",
					},
					Shell: &detector.ShellInfo{
						Name:      detector.ShellTypeZsh,
						Version:   "5.9",
						IsDefault: true,
					},
					Terminal: &detector.TerminalInfo{
						Name:    "kitty",
						Version: "0.30",
					},
					Fonts: &detector.FontInventory{
						NerdFonts: []detector.FontInfo{
							{Name: "FiraCode Nerd Font"},
						},
					},
					ExistingConfigs: &detector.ConfigSnapshot{
						HasOhMyPosh: true,
						HasStarship: false,
					},
				},
			},
			wantContains: []string{
				"System Detection",
				"OS:",
				"Shell:",
				"Terminal:",
				"Fonts:",
				"Installed Components",
			},
		},
		{
			name: "loading state",
			model: Model{
				CurrentScreen: ScreenDetect,
				Loading:       true,
			},
			wantContains: []string{
				"System Detection",
			},
		},
		{
			name: "empty detection result",
			model: Model{
				CurrentScreen:  ScreenDetect,
				systemInfo:     &SystemInfo{},
				DetectorResult: nil,
			},
			wantContains: []string{
				"System Detection",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.model.renderDetect()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("renderDetect() missing %q in output", want)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("renderDetect() should not contain %q", notWant)
				}
			}
		})
	}
}

func TestRenderThemeSelect(t *testing.T) {
	m := NewModel()
	m.CurrentScreen = ScreenThemeSelect
	m.Items = []string{"dark", "light", "solarized"}
	m.Cursor = 1
	m.Width = 80
	m.Height = 24

	output := m.renderThemeSelect()

	// Check for title
	if !strings.Contains(output, "Select Theme") {
		t.Error("renderThemeSelect() missing title")
	}

	// Check for items
	for _, item := range m.Items {
		if !strings.Contains(output, item) {
			t.Errorf("renderThemeSelect() missing item %q", item)
		}
	}

	// Check for cursor indicator
	if !strings.Contains(output, "→") {
		t.Error("renderThemeSelect() missing cursor indicator")
	}
}

func TestRenderFontSelect(t *testing.T) {
	m := NewModel()
	m.CurrentScreen = ScreenFontSelect
	m.Items = []string{"FiraCode", "JetBrainsMono", "Hack"}
	m.Cursor = 0

	output := m.renderFontSelect()

	// Check for title
	if !strings.Contains(output, "Select Font") {
		t.Error("renderFontSelect() missing title")
	}

	// Check for items
	for _, item := range m.Items {
		if !strings.Contains(output, item) {
			t.Errorf("renderFontSelect() missing item %q", item)
		}
	}
}

func TestRenderPreview(t *testing.T) {
	m := NewModel()
	m.CurrentScreen = ScreenPreview

	output := m.renderPreview()

	wantContains := []string{
		"Preview",
		"Press Enter to install",
	}

	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("renderPreview() missing %q in output", want)
		}
	}
}

func TestRenderInstall(t *testing.T) {
	tests := []struct {
		name         string
		loading      bool
		loadingMsg   string
		wantContains []string
	}{
		{
			name:         "loading state",
			loading:      true,
			loadingMsg:   "Installing components...",
			wantContains: []string{"Installing", "Installing components..."},
		},
		{
			name:         "complete state",
			loading:      false,
			wantContains: []string{"Installing", "Installation complete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				CurrentScreen:  ScreenInstall,
				Loading:        tt.loading,
				LoadingMessage: tt.loadingMsg,
			}

			output := m.renderInstall()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("renderInstall() missing %q in output", want)
				}
			}
		})
	}
}

func TestRenderComplete(t *testing.T) {
	tests := []struct {
		name         string
		systemInfo   *SystemInfo
		wantContains []string
	}{
		{
			name: "with system info",
			systemInfo: &SystemInfo{
				OS:       "Linux",
				Shell:    "zsh",
				Terminal: "kitty",
			},
			wantContains: []string{
				"Complete",
				"configured successfully",
				"Restart your shell",
			},
		},
		{
			name:         "without system info",
			systemInfo:   nil,
			wantContains: []string{"Complete", "configured successfully"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				CurrentScreen: ScreenComplete,
				systemInfo:    tt.systemInfo,
			}

			output := m.renderComplete()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("renderComplete() missing %q in output", want)
				}
			}
		})
	}
}

func TestRenderError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantContains []string
	}{
		{
			name:         "with error",
			err:          assertError("test error"),
			wantContains: []string{"Error", "test error", "Press Esc"},
		},
		{
			name:         "without error",
			err:          nil,
			wantContains: []string{"Error", "unknown error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				CurrentScreen: ScreenError,
				Error:         tt.err,
			}

			output := m.renderError()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("renderError() missing %q in output", want)
				}
			}
		})
	}
}

func TestRenderFooter(t *testing.T) {
	tests := []struct {
		screen       Screen
		wantContains []string
	}{
		{ScreenWelcome, []string{"Enter", "Continue"}},
		{ScreenDetect, []string{"Enter", "Continue", "Esc", "Back"}},
		{ScreenThemeSelect, []string{"Navigate", "Select", "Back"}},
		{ScreenFontSelect, []string{"Navigate", "Select", "Back"}},
		{ScreenPreview, []string{"Enter", "Install", "Esc", "Back"}},
		{ScreenComplete, []string{"Enter", "Finish"}},
		{ScreenError, []string{"Esc", "Back"}},
	}

	for _, tt := range tests {
		t.Run(tt.screen.String(), func(t *testing.T) {
			m := Model{CurrentScreen: tt.screen}
			output := m.renderFooter()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("renderFooter() for %s missing %q", tt.screen, want)
				}
			}
		})
	}
}

func TestView(t *testing.T) {
	tests := []struct {
		screen       Screen
		loading      bool
		wantContains []string
	}{
		{ScreenWelcome, false, []string{"Welcome"}},
		{ScreenDetect, false, []string{"System Detection"}},
		{ScreenThemeSelect, false, []string{"Select Theme"}},
		{ScreenFontSelect, false, []string{"Select Font"}},
		{ScreenPreview, false, []string{"Preview"}},
		{ScreenInstall, false, []string{"Installing"}},
		{ScreenComplete, false, []string{"Complete"}},
		{ScreenError, false, []string{"Error"}},
	}

	for _, tt := range tests {
		t.Run(tt.screen.String(), func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = tt.screen
			m.Loading = tt.loading
			if tt.screen == ScreenThemeSelect || tt.screen == ScreenFontSelect {
				m.Items = []string{"item1", "item2"}
			}

			output := m.View()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("View() for %s missing %q", tt.screen, want)
				}
			}
		})
	}
}

func TestViewLoading(t *testing.T) {
	m := NewModel()
	m.Loading = true
	m.LoadingMessage = "Processing..."

	output := m.View()

	if !strings.Contains(output, "Processing...") {
		t.Error("View() should show loading message when Loading is true")
	}
}

func TestFormatSystemInfoEmpty(t *testing.T) {
	info := formatSystemInfo(nil)

	if info.OS != "" || info.Shell != "" || info.Terminal != "" {
		t.Error("formatSystemInfo(nil) should return empty SystemInfo")
	}
}

func TestFormatSystemInfoFull(t *testing.T) {
	result := &detector.DetectorResult{
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
			Name:    "iTerm2",
			Version: "3.4",
		},
		Fonts: &detector.FontInventory{
			NerdFonts: []detector.FontInfo{
				{Name: "FiraCode Nerd Font"},
				{Name: "JetBrainsMono Nerd Font"},
			},
		},
	}

	info := formatSystemInfo(result)

	if info.OS == "" {
		t.Error("OS should not be empty")
	}

	if info.Shell == "" {
		t.Error("Shell should not be empty")
	}

	if info.Terminal == "" {
		t.Error("Terminal should not be empty")
	}

	if len(info.Fonts) != 2 {
		t.Errorf("Expected 2 fonts, got %d", len(info.Fonts))
	}
}

func TestJoinHelpers(t *testing.T) {
	// Test JoinHorizontal
	result := JoinHorizontal("a", "b", "c")
	if result == "" {
		t.Error("JoinHorizontal should return non-empty string")
	}

	// Test JoinVertical
	result = JoinVertical("a", "b", "c")
	if result == "" {
		t.Error("JoinVertical should return non-empty string")
	}
}

func TestFormatStatus(t *testing.T) {
	// Test with success=true
	result := FormatStatus("Label", "Value", true)
	if !strings.Contains(result, "Label") {
		t.Error("FormatStatus should contain label")
	}
	if !strings.Contains(result, "Value") {
		t.Error("FormatStatus should contain value")
	}

	// Test with success=false
	result = FormatStatus("Label", "Value", false)
	if !strings.Contains(result, "Label") {
		t.Error("FormatStatus should contain label (warning)")
	}
}

// assertError creates a simple error for testing
func assertError(msg string) error {
	return &testError{msg: msg}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
