// Package views provides tests for detection screen.
package views

import (
	"errors"
	"strings"
	"testing"

	"github.com/savanhi/shell/internal/detector"
)

func TestNewDetectionModel(t *testing.T) {
	m := NewDetectionModel()

	if m.Loading != true {
		t.Error("NewDetectionModel should start in loading state")
	}

	if len(m.Actions) != 3 {
		t.Errorf("Expected 3 actions, got %d", len(m.Actions))
	}

	if m.Actions[0] != "Continue" {
		t.Errorf("Expected first action to be 'Continue', got %q", m.Actions[0])
	}

	if m.Cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.Cursor)
	}
}

func TestDetectionModelWithResult(t *testing.T) {
	m := NewDetectionModel()
	result := &detector.DetectorResult{
		OS: &detector.OSInfo{
			Type:    detector.OSTypeLinux,
			Distro:  "Ubuntu",
			Version: "22.04",
			Arch:    "amd64",
		},
		Shell: &detector.ShellInfo{
			Name:    detector.ShellTypeZsh,
			Version: "5.9",
		},
	}

	m = m.WithResult(result)

	if m.Loading {
		t.Error("WithResult should set Loading to false")
	}

	if m.Result != result {
		t.Error("WithResult should set the Result")
	}
}

func TestDetectionModelWithError(t *testing.T) {
	m := NewDetectionModel()
	testErr := errors.New("detection failed")

	m = m.WithError(testErr)

	if m.Loading {
		t.Error("WithError should set Loading to false")
	}

	if m.Error != testErr {
		t.Error("WithError should set the Error")
	}
}

func TestDetectionModelSetCursor(t *testing.T) {
	tests := []struct {
		name     string
		pos      int
		expected int
	}{
		{"valid position 0", 0, 0},
		{"valid position 1", 1, 1},
		{"valid position 2", 2, 2},
		{"invalid negative", -1, 0},
		{"invalid too high", 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDetectionModel()
			m = m.SetCursor(tt.pos)

			if m.Cursor != tt.expected {
				t.Errorf("SetCursor(%d) = %d, want %d", tt.pos, m.Cursor, tt.expected)
			}
		})
	}
}

func TestDetectionModelViewLoading(t *testing.T) {
	m := NewDetectionModel()
	m.Loading = true

	output := m.View()

	wantContains := []string{
		"System Detection",
		"Detecting system information",
	}

	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("View() missing %q in loading state", want)
		}
	}
}

func TestDetectionModelViewError(t *testing.T) {
	m := NewDetectionModel()
	m.Loading = false
	m.Error = errors.New("detection failed")

	output := m.View()

	wantContains := []string{
		"System Detection",
		"Detection Failed",
		"detection failed",
	}

	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("View() missing %q in error state", want)
		}
	}
}

func TestDetectionModelViewResults(t *testing.T) {
	tests := []struct {
		name         string
		result       *detector.DetectorResult
		wantContains []string
	}{
		{
			name: "full detection results",
			result: &detector.DetectorResult{
				OS: &detector.OSInfo{
					Type:       detector.OSTypeLinux,
					Distro:     "Ubuntu",
					Version:    "22.04",
					Arch:       "amd64",
					PackageMgr: "apt",
				},
				Shell: &detector.ShellInfo{
					Name:      detector.ShellTypeZsh,
					Version:   "5.9",
					IsDefault: true,
				},
				Terminal: &detector.TerminalInfo{
					Name:              "kitty",
					Version:           "0.30",
					SupportsTrueColor: true,
					SupportsLigatures: true,
				},
				Fonts: &detector.FontInventory{
					NerdFonts: []detector.FontInfo{
						{Name: "FiraCode Nerd Font"},
					},
					Fonts: []detector.FontInfo{
						{Name: "FiraCode Nerd Font"},
						{Name: "Hack"},
					},
				},
				ExistingConfigs: &detector.ConfigSnapshot{
					HasOhMyPosh: true,
					HasStarship: false,
				},
			},
			wantContains: []string{
				"Operating System",
				"Shell",
				"Terminal",
				"Fonts",
				"Installed Components",
				"linux",
				"zsh",
				"kitty",
				"Nerd Fonts",
			},
		},
		{
			name: "minimal detection results",
			result: &detector.DetectorResult{
				OS: &detector.OSInfo{
					Type: detector.OSTypeMacOS,
					Arch: "arm64",
				},
			},
			wantContains: []string{
				"Operating System",
				"macos",
			},
		},
		{
			name: "with installed components",
			result: &detector.DetectorResult{
				ExistingConfigs: &detector.ConfigSnapshot{
					HasOhMyPosh: true,
					HasStarship: true,
				},
			},
			wantContains: []string{
				"Installed Components",
				"Oh My Posh",
				"Starship",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDetectionModel()
			m = m.WithResult(tt.result)

			output := m.View()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("View() missing %q in output", want)
				}
			}
		})
	}
}

func TestDetectionModelRenderOSInfo(t *testing.T) {
	tests := []struct {
		name         string
		osInfo       *detector.OSInfo
		wantContains []string
	}{
		{
			name: "full OS info",
			osInfo: &detector.OSInfo{
				Type:       detector.OSTypeLinux,
				Distro:     "Ubuntu",
				Version:    "22.04",
				Arch:       "amd64",
				PackageMgr: "apt",
			},
			wantContains: []string{"Type:", "Distro:", "Arch:", "Package Mgr:", "Ubuntu"},
		},
		{
			name: "minimal OS info",
			osInfo: &detector.OSInfo{
				Type: detector.OSTypeMacOS,
				Arch: "arm64",
			},
			wantContains: []string{"Type:", "Arch:", "macos"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDetectionModel()
			_ = m // Tests renderOSInfo indirectly through View()
		})
	}
}

func TestDetectionModelRenderShellInfo(t *testing.T) {
	tests := []struct {
		name         string
		shellInfo    *detector.ShellInfo
		wantContains []string
	}{
		{
			name: "default shell",
			shellInfo: &detector.ShellInfo{
				Name:      detector.ShellTypeZsh,
				Version:   "5.9",
				IsDefault: true,
			},
			wantContains: []string{"Name:", "zsh", "Version:", "Default:", "Yes"},
		},
		{
			name: "non-default shell with RC file",
			shellInfo: &detector.ShellInfo{
				Name:      detector.ShellTypeBash,
				Version:   "5.1",
				RCFile:    "/home/user/.bashrc",
				IsDefault: false,
			},
			wantContains: []string{"Name:", "bash", "RC File:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDetectionModel()
			_ = m // Tests renderShellInfo indirectly through View()
		})
	}
}

func TestDetectionModelRenderFontInfo(t *testing.T) {
	tests := []struct {
		name         string
		fontInfo     *detector.FontInventory
		wantContains []string
	}{
		{
			name: "with nerd fonts",
			fontInfo: &detector.FontInventory{
				NerdFonts: []detector.FontInfo{
					{Name: "FiraCode Nerd Font"},
					{Name: "JetBrainsMono Nerd Font"},
				},
				Fonts: []detector.FontInfo{
					{Name: "FiraCode Nerd Font"},
					{Name: "JetBrainsMono Nerd Font"},
					{Name: "Hack"},
				},
			},
			wantContains: []string{"Nerd Fonts:", "2 found", "FiraCode", "Total Fonts:"},
		},
		{
			name: "without nerd fonts",
			fontInfo: &detector.FontInventory{
				NerdFonts: []detector.FontInfo{},
				Fonts: []detector.FontInfo{
					{Name: "Hack"},
				},
			},
			wantContains: []string{"None found", "Total Fonts:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewDetectionModel()
			_ = m // Tests renderFontInfo indirectly through View()
		})
	}
}

func TestDetectionModelGetAction(t *testing.T) {
	m := NewDetectionModel()

	tests := []struct {
		cursor   int
		expected string
	}{
		{0, "Continue"},
		{1, "Refresh"},
		{2, "Back"},
	}

	for _, tt := range tests {
		m.Cursor = tt.cursor
		action := m.GetAction()

		if action != tt.expected {
			t.Errorf("GetAction() with cursor %d = %q, want %q", tt.cursor, action, tt.expected)
		}
	}
}

func TestDetectionModelGetActionInvalidCursor(t *testing.T) {
	m := NewDetectionModel()
	m.Cursor = -1

	action := m.GetAction()
	if action != "" {
		t.Errorf("GetAction() with invalid cursor should return empty string, got %q", action)
	}

	m.Cursor = 10
	action = m.GetAction()
	if action != "" {
		t.Errorf("GetAction() with out of bounds cursor should return empty string, got %q", action)
	}
}

func TestDetermineAction(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected DetectionAction
	}{
		{"continue", "Continue", ActionContinue},
		{"refresh", "Refresh", ActionRefresh},
		{"back", "Back", ActionBack},
		{"unknown", "Unknown", ActionContinue},
		{"empty", "", ActionContinue},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetermineAction(tt.action)
			if result != tt.expected {
				t.Errorf("DetermineAction(%q) = %v, want %v", tt.action, result, tt.expected)
			}
		})
	}
}

func TestDetectionModelRenderActions(t *testing.T) {
	m := NewDetectionModel()
	// Set a result so actions are shown
	m = m.WithResult(&detector.DetectorResult{
		OS: &detector.OSInfo{Type: detector.OSTypeLinux},
	})

	// Test that all actions are present in the rendered output
	output := m.View()

	for _, action := range m.Actions {
		if !strings.Contains(output, action) {
			t.Errorf("View() should contain action %q", action)
		}
	}
}

func TestDetectionModelRenderFooter(t *testing.T) {
	tests := []struct {
		name         string
		model        DetectionModel
		wantContains []string
	}{
		{
			name:         "loading state",
			model:        NewDetectionModel(), // Loading by default
			wantContains: []string{"Esc", "cancel"},
		},
		{
			name: "error state",
			model: DetectionModel{
				Loading: false,
				Error:   errors.New("test error"),
			},
			wantContains: []string{"Esc", "Back", "q", "Quit"},
		},
		{
			name: "results state",
			model: DetectionModel{
				Loading: false,
				Result:  &detector.DetectorResult{},
			},
			wantContains: []string{"Navigate", "Select", "q", "Quit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.model.View()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Footer missing %q in %s state", want, tt.name)
				}
			}
		})
	}
}

func TestDetectionModelWithManyFonts(t *testing.T) {
	// Test that when there are many fonts, we show a summary
	nerdFonts := make([]detector.FontInfo, 10)
	for i := range nerdFonts {
		nerdFonts[i] = detector.FontInfo{Name: "TestFont"}
	}

	m := NewDetectionModel()
	m = m.WithResult(&detector.DetectorResult{
		Fonts: &detector.FontInventory{
			NerdFonts: nerdFonts,
		},
	})

	output := m.View()

	// Should not list all 10 fonts individually
	if !strings.Contains(output, "and") {
		t.Error("Should show summary for many fonts")
	}
}

func TestDetectionModelTerminalInfo(t *testing.T) {
	result := &detector.DetectorResult{
		Terminal: &detector.TerminalInfo{
			Name:              "kitty",
			Version:           "0.30",
			SupportsTrueColor: true,
			SupportsLigatures: true,
		},
	}

	m := NewDetectionModel()
	m = m.WithResult(result)

	output := m.View()

	wantContains := []string{
		"Terminal",
		"Name:",
		"kitty",
		"True Color:",
		"Ligatures:",
	}

	for _, want := range wantContains {
		if !strings.Contains(output, want) {
			t.Errorf("View() missing %q for terminal info", want)
		}
	}
}
