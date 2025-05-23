package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines a set of keybindings.
type KeyMap struct {
	Quit   key.Binding
	Submit key.Binding
}

// NewKeyMap returns a set of keybindings.
func NewKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("ESC", "quit"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("ENTER", "submit"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return nil
}
