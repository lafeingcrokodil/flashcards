package tui

import "github.com/charmbracelet/lipgloss"

var lightBlue = lipgloss.Color("12")
var red = lipgloss.Color("9")

var contextStyle = lipgloss.NewStyle().Italic(true).Foreground(lightBlue)
var expectedStyle = lipgloss.NewStyle().Foreground(red)
