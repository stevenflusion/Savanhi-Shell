// Package tui provides keybindings for the TUI.
package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// KeyMap defines the keybindings for the TUI.
type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Left     key.Binding
	Right    key.Binding
	Enter    key.Binding
	Escape   key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Help     key.Binding
	Quit     key.Binding
	Confirm  key.Binding
	Cancel   key.Binding
	Select   key.Binding
	All      key.Binding
	None     key.Binding
	Space    key.Binding
}

// DefaultKeyMap returns the default keybindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "go back"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "cancel"),
		),
		Select: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "select/deselect"),
		),
		All: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		None: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "deselect all"),
		),
		Space: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select/toggle"),
		),
	}
}

// Keys is the global keybindings instance.
var Keys = DefaultKeyMap()

// ShortHelp returns a short help message.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Escape, k.Quit}
}

// FullHelp returns a full help message.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Escape, k.Tab, k.ShiftTab},
		{k.Help, k.Quit, k.Confirm, k.Cancel},
		{k.Select, k.All, k.None},
	}
}

// IsNavigationKey checks if the key is a navigation key.
func IsNavigationKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, Keys.Up, Keys.Down, Keys.Left, Keys.Right)
}

// IsSelectionKey checks if the key is a selection key.
func IsSelectionKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, Keys.Enter, Keys.Space)
}

// IsCancelKey checks if the key is a cancel key.
func IsCancelKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, Keys.Escape, Keys.Cancel)
}

// IsQuitKey checks if the key is a quit key.
func IsQuitKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, Keys.Quit)
}

// IsConfirmKey checks if the key is a confirm key.
func IsConfirmKey(msg tea.KeyMsg) bool {
	return key.Matches(msg, Keys.Confirm, Keys.Enter)
}
