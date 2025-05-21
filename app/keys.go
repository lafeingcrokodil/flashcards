package app

import "github.com/charmbracelet/bubbles/key"

// keyMap defines a set of keybindings.
type keyMap struct {
	Quit   key.Binding
	Submit key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Submit, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return nil
}
