// Package tui provides tests for TUI update logic.
package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleKeyPressQuit(t *testing.T) {
	tests := []struct {
		name     string
		screen   Screen
		key      tea.KeyMsg
		wantQuit bool
	}{
		{
			name:     "q key quits from welcome",
			screen:   ScreenWelcome,
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantQuit: true,
		},
		{
			name:     "ctrl+c quits from welcome",
			screen:   ScreenWelcome,
			key:      tea.KeyMsg{Type: tea.KeyCtrlC},
			wantQuit: true,
		},
		{
			name:     "q key does not quit when enter pressed",
			screen:   ScreenWelcome,
			key:      tea.KeyMsg{Type: tea.KeyEnter},
			wantQuit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = tt.screen

			newModel, cmd := m.Update(tt.key)
			m = newModel.(Model)

			if tt.wantQuit {
				if !m.Quitting {
					t.Error("expected model to be quitting")
				}
				if cmd == nil {
					t.Error("expected quit command")
				}
			} else {
				if m.Quitting {
					t.Error("expected model to not be quitting")
				}
			}
		})
	}
}

func TestHandleWelcomeKeys(t *testing.T) {
	tests := []struct {
		name           string
		key            tea.KeyMsg
		wantScreen     Screen
		wantLoading    bool
		wantLoadingMsg string
	}{
		{
			name:           "enter key moves to detect screen",
			key:            tea.KeyMsg{Type: tea.KeyEnter},
			wantScreen:     ScreenDetect,
			wantLoading:    true,
			wantLoadingMsg: "Detecting system information...",
		},
		{
			name:        "space key moves to detect screen",
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}},
			wantScreen:  ScreenDetect,
			wantLoading: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = ScreenWelcome

			newModel, _ := m.Update(tt.key)
			m = newModel.(Model)

			if m.CurrentScreen != tt.wantScreen {
				t.Errorf("expected screen %v, got %v", tt.wantScreen, m.CurrentScreen)
			}

			if m.Loading != tt.wantLoading {
				t.Errorf("expected Loading=%v, got %v", tt.wantLoading, m.Loading)
			}

			if tt.wantLoadingMsg != "" && m.LoadingMessage != tt.wantLoadingMsg {
				t.Errorf("expected LoadingMessage=%q, got %q", tt.wantLoadingMsg, m.LoadingMessage)
			}
		})
	}
}

func TestHandleDetectKeys(t *testing.T) {
	m := NewModel()
	m.CurrentScreen = ScreenDetect

	// Test enter key moves to theme select
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	if m.CurrentScreen != ScreenThemeSelect {
		t.Errorf("expected screen ThemeSelect, got %v", m.CurrentScreen)
	}

	if cmd != nil {
		t.Error("expected no command from ThemeSelect transition")
	}
}

func TestHandleThemeSelectKeysNavigation(t *testing.T) {
	tests := []struct {
		name        string
		startCursor int
		key         tea.KeyMsg
		endCursor   int
		numItems    int
	}{
		{
			name:        "down moves cursor",
			startCursor: 0,
			key:         tea.KeyMsg{Type: tea.KeyDown},
			endCursor:   1,
			numItems:    5,
		},
		{
			name:        "up moves cursor",
			startCursor: 1,
			key:         tea.KeyMsg{Type: tea.KeyUp},
			endCursor:   0,
			numItems:    5,
		},
		{
			name:        "down at bottom stays",
			startCursor: 4,
			key:         tea.KeyMsg{Type: tea.KeyDown},
			endCursor:   4,
			numItems:    5,
		},
		{
			name:        "up at top stays",
			startCursor: 0,
			key:         tea.KeyMsg{Type: tea.KeyUp},
			endCursor:   0,
			numItems:    5,
		},
		{
			name:        "k moves up",
			startCursor: 1,
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}},
			endCursor:   0,
			numItems:    5,
		},
		{
			name:        "j moves down",
			startCursor: 0,
			key:         tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}},
			endCursor:   1,
			numItems:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = ScreenThemeSelect
			m.Cursor = tt.startCursor
			m.Items = []string{"item1", "item2", "item3", "item4", "item5"}

			newModel, _ := m.Update(tt.key)
			m = newModel.(Model)

			if m.Cursor != tt.endCursor {
				t.Errorf("cursor = %d, want %d", m.Cursor, tt.endCursor)
			}
		})
	}
}

func TestHandleThemeSelectKeysActions(t *testing.T) {
	tests := []struct {
		name         string
		key          tea.KeyMsg
		cursor       int
		items        []string
		wantScreen   Screen
		wantCursor   int
		wantSelected bool
	}{
		{
			name:         "enter selects and moves to font",
			key:          tea.KeyMsg{Type: tea.KeyEnter},
			cursor:       0,
			items:        []string{"theme1", "theme2"},
			wantScreen:   ScreenFontSelect,
			wantCursor:   0,
			wantSelected: true,
		},
		{
			name:         "space selects and moves to font",
			key:          tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}},
			cursor:       1,
			items:        []string{"theme1", "theme2"},
			wantScreen:   ScreenFontSelect,
			wantCursor:   0,
			wantSelected: true,
		},
		{
			name:       "escape goes back to detect",
			key:        tea.KeyMsg{Type: tea.KeyEsc},
			cursor:     1,
			items:      []string{"theme1", "theme2"},
			wantScreen: ScreenDetect,
			wantCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = ScreenThemeSelect
			m.Cursor = tt.cursor
			m.Items = tt.items

			newModel, _ := m.Update(tt.key)
			m = newModel.(Model)

			if m.CurrentScreen != tt.wantScreen {
				t.Errorf("expected screen %v, got %v", tt.wantScreen, m.CurrentScreen)
			}

			if m.Cursor != tt.wantCursor {
				t.Errorf("expected cursor %d, got %d", tt.wantCursor, m.Cursor)
			}

			if tt.wantSelected {
				if !m.Selected["theme"] {
					t.Error("expected theme to be selected")
				}
			}
		})
	}
}

func TestHandleFontSelectKeys(t *testing.T) {
	tests := []struct {
		name         string
		key          tea.KeyMsg
		cursor       int
		items        []string
		wantScreen   Screen
		wantCursor   int
		wantSelected bool
	}{
		{
			name:         "enter selects and moves to preview",
			key:          tea.KeyMsg{Type: tea.KeyEnter},
			cursor:       0,
			items:        []string{"font1", "font2"},
			wantScreen:   ScreenPreview,
			wantCursor:   0,
			wantSelected: true,
		},
		{
			name:       "escape goes back to theme select",
			key:        tea.KeyMsg{Type: tea.KeyEsc},
			cursor:     1,
			items:      []string{"font1", "font2"},
			wantScreen: ScreenThemeSelect,
			wantCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = ScreenFontSelect
			m.Cursor = tt.cursor
			m.Items = tt.items
			m.Selected = make(map[string]bool)

			newModel, _ := m.Update(tt.key)
			m = newModel.(Model)

			if m.CurrentScreen != tt.wantScreen {
				t.Errorf("expected screen %v, got %v", tt.wantScreen, m.CurrentScreen)
			}

			if m.Cursor != tt.wantCursor {
				t.Errorf("expected cursor %d, got %d", tt.wantCursor, m.Cursor)
			}

			if tt.wantSelected {
				if !m.Selected["font"] {
					t.Error("expected font to be selected")
				}
			}
		})
	}
}

func TestHandlePreviewKeys(t *testing.T) {
	tests := []struct {
		name       string
		key        tea.KeyMsg
		wantScreen Screen
	}{
		{
			name:       "enter moves to install",
			key:        tea.KeyMsg{Type: tea.KeyEnter},
			wantScreen: ScreenInstall,
		},
		{
			name:       "escape goes back to font select",
			key:        tea.KeyMsg{Type: tea.KeyEsc},
			wantScreen: ScreenFontSelect,
		},
		{
			name:       "y confirms and moves to install",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
			wantScreen: ScreenInstall,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = ScreenPreview

			newModel, _ := m.Update(tt.key)
			m = newModel.(Model)

			if m.CurrentScreen != tt.wantScreen {
				t.Errorf("expected screen %v, got %v", tt.wantScreen, m.CurrentScreen)
			}
		})
	}
}

func TestHandleInstallKeys(t *testing.T) {
	tests := []struct {
		name       string
		loading    bool
		key        tea.KeyMsg
		wantScreen Screen
	}{
		{
			name:       "enter when not loading goes to complete",
			loading:    false,
			key:        tea.KeyMsg{Type: tea.KeyEnter},
			wantScreen: ScreenComplete,
		},
		{
			name:       "escape goes back to preview",
			loading:    false,
			key:        tea.KeyMsg{Type: tea.KeyEsc},
			wantScreen: ScreenPreview,
		},
		{
			name:       "escape during loading goes back to preview",
			loading:    true,
			key:        tea.KeyMsg{Type: tea.KeyEsc},
			wantScreen: ScreenPreview,
		},
		{
			name:       "enter during loading does nothing",
			loading:    true,
			key:        tea.KeyMsg{Type: tea.KeyEnter},
			wantScreen: ScreenInstall,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = ScreenInstall
			m.Loading = tt.loading

			newModel, _ := m.Update(tt.key)
			m = newModel.(Model)

			if m.CurrentScreen != tt.wantScreen {
				t.Errorf("expected screen %v, got %v", tt.wantScreen, m.CurrentScreen)
			}
		})
	}
}

func TestHandleCompleteKeys(t *testing.T) {
	tests := []struct {
		name     string
		key      tea.KeyMsg
		wantQuit bool
	}{
		{
			name:     "enter quits",
			key:      tea.KeyMsg{Type: tea.KeyEnter},
			wantQuit: true,
		},
		{
			name:     "q quits",
			key:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantQuit: true,
		},
		{
			name:     "ctrl+c quits",
			key:      tea.KeyMsg{Type: tea.KeyCtrlC},
			wantQuit: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = ScreenComplete

			newModel, cmd := m.Update(tt.key)
			m = newModel.(Model)

			if m.Quitting != tt.wantQuit {
				t.Errorf("expected Quitting=%v, got %v", tt.wantQuit, m.Quitting)
			}

			if tt.wantQuit && cmd == nil {
				t.Error("expected quit command")
			}
		})
	}
}

func TestHandleErrorKeys(t *testing.T) {
	tests := []struct {
		name       string
		key        tea.KeyMsg
		wantScreen Screen
		wantError  bool
	}{
		{
			name:       "escape clears error and goes to welcome",
			key:        tea.KeyMsg{Type: tea.KeyEsc},
			wantScreen: ScreenWelcome,
			wantError:  false,
		},
		{
			name:       "q quits from error",
			key:        tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantScreen: ScreenError,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel()
			m.CurrentScreen = ScreenError
			m.Error = errors.New("test error")

			if tt.key.Type == tea.KeyRunes && len(tt.key.Runes) > 0 && tt.key.Runes[0] == 'q' {
				newModel, cmd := m.Update(tt.key)
				m = newModel.(Model)

				if !m.Quitting {
					t.Error("expected Quitting to be true")
				}
				if cmd == nil {
					t.Error("expected quit command")
				}
			} else {
				newModel, _ := m.Update(tt.key)
				m = newModel.(Model)

				if m.CurrentScreen != tt.wantScreen {
					t.Errorf("expected screen %v, got %v", tt.wantScreen, m.CurrentScreen)
				}

				if tt.wantError {
					if m.Error == nil {
						t.Error("expected error to remain")
					}
				} else {
					if m.Error != nil {
						t.Error("expected error to be cleared")
					}
				}
			}
		})
	}
}

func TestKeyBindingFunctions(t *testing.T) {
	// Test IsNavigationKey
	t.Run("IsNavigationKey", func(t *testing.T) {
		tests := []struct {
			key      tea.KeyMsg
			expected bool
		}{
			{tea.KeyMsg{Type: tea.KeyUp}, true},
			{tea.KeyMsg{Type: tea.KeyDown}, true},
			{tea.KeyMsg{Type: tea.KeyLeft}, true},
			{tea.KeyMsg{Type: tea.KeyRight}, true},
			{tea.KeyMsg{Type: tea.KeyEnter}, false},
			{tea.KeyMsg{Type: tea.KeyEsc}, false},
		}

		for _, tt := range tests {
			result := IsNavigationKey(tt.key)
			if result != tt.expected {
				t.Errorf("IsNavigationKey(%v) = %v, want %v", tt.key, result, tt.expected)
			}
		}
	})

	// Test IsSelectionKey
	t.Run("IsSelectionKey", func(t *testing.T) {
		tests := []struct {
			key      tea.KeyMsg
			expected bool
		}{
			{tea.KeyMsg{Type: tea.KeyEnter}, true},
			{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}, true},
			{tea.KeyMsg{Type: tea.KeyUp}, false},
			{tea.KeyMsg{Type: tea.KeyEsc}, false},
		}

		for _, tt := range tests {
			result := IsSelectionKey(tt.key)
			if result != tt.expected {
				t.Errorf("IsSelectionKey(%v) = %v, want %v", tt.key, result, tt.expected)
			}
		}
	})

	// Test IsCancelKey
	t.Run("IsCancelKey", func(t *testing.T) {
		tests := []struct {
			key      tea.KeyMsg
			expected bool
		}{
			{tea.KeyMsg{Type: tea.KeyEsc}, true},
			{tea.KeyMsg{Type: tea.KeyEnter}, false},
			{tea.KeyMsg{Type: tea.KeyUp}, false},
		}

		for _, tt := range tests {
			result := IsCancelKey(tt.key)
			if result != tt.expected {
				t.Errorf("IsCancelKey(%v) = %v, want %v", tt.key, result, tt.expected)
			}
		}
	})

	// Test IsQuitKey
	t.Run("IsQuitKey", func(t *testing.T) {
		tests := []struct {
			key      tea.KeyMsg
			expected bool
		}{
			{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}, true},
			{tea.KeyMsg{Type: tea.KeyCtrlC}, true},
			{tea.KeyMsg{Type: tea.KeyEnter}, false},
			{tea.KeyMsg{Type: tea.KeyEsc}, false},
		}

		for _, tt := range tests {
			result := IsQuitKey(tt.key)
			if result != tt.expected {
				t.Errorf("IsQuitKey(%v) = %v, want %v", tt.key, result, tt.expected)
			}
		}
	})

	// Test IsConfirmKey
	t.Run("IsConfirmKey", func(t *testing.T) {
		tests := []struct {
			key      tea.KeyMsg
			expected bool
		}{
			{tea.KeyMsg{Type: tea.KeyEnter}, true},
			{tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}, true},
			{tea.KeyMsg{Type: tea.KeyUp}, false},
			{tea.KeyMsg{Type: tea.KeyEsc}, false},
		}

		for _, tt := range tests {
			result := IsConfirmKey(tt.key)
			if result != tt.expected {
				t.Errorf("IsConfirmKey(%v) = %v, want %v", tt.key, result, tt.expected)
			}
		}
	})
}

func TestKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	// Test that all key bindings are defined by checking their help text
	if km.Up.Help().Desc == "" {
		t.Error("Up key binding not defined")
	}
	if km.Down.Help().Desc == "" {
		t.Error("Down key binding not defined")
	}
	if km.Enter.Help().Desc == "" {
		t.Error("Enter key binding not defined")
	}
	if km.Escape.Help().Desc == "" {
		t.Error("Escape key binding not defined")
	}
	if km.Quit.Help().Desc == "" {
		t.Error("Quit key binding not defined")
	}
}

func TestKeyMapHelp(t *testing.T) {
	km := DefaultKeyMap()

	shortHelp := km.ShortHelp()
	if len(shortHelp) == 0 {
		t.Error("ShortHelp should return at least one key binding")
	}

	fullHelp := km.FullHelp()
	if len(fullHelp) == 0 {
		t.Error("FullHelp should return at least one row of key bindings")
	}
}

func TestDetectSystemFunction(t *testing.T) {
	// The detectSystem function is a tea.Cmd, we can at least verify it compiles
	// and returns the correct type
	cmd := detectSystem()
	if cmd == nil {
		t.Error("detectSystem() should return a non-nil command")
	}
}
