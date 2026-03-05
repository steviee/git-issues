package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the key bindings for the board.
type KeyMap struct {
	Left      key.Binding
	Right     key.Binding
	Up        key.Binding
	Down      key.Binding
	MoveLeft  key.Binding
	MoveRight key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Quit      key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("h/←", "column"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("l/→", "column"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		MoveLeft: key.NewBinding(
			key.WithKeys("H", "shift+left"),
			key.WithHelp("H", "move left"),
		),
		MoveRight: key.NewBinding(
			key.WithKeys("L", "shift+right"),
			key.WithHelp("L", "move right"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("Enter", "details"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("Esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
