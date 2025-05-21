package app

import (
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// TUI is a terminal user interface for reviewing flash cards.
type TUI struct {
	// LoadConfig configures how to load flashcards from a file.
	LoadConfig LoadConfig
	// flashcards is the list of flashcards to be reviewed.
	flashcards []*Flashcard
	// cursor is the index of the current flashcard being reviewed.
	cursor int
	// viewCount is the number of flashcards that have been reviewed.
	viewCount int
	// correctCount is the number of correct answers so far.
	correctCount int
	// showExpected is true if the expected answer should be shown.
	showExpected bool
	// answer is the user-provided answer.
	answer textinput.Model
	// help displays keybindings to the user.
	help help.Model
	// keys are the keybindings used by the UI.
	keys keyMap
}

// NewTUI returns a new TUI.
func NewTUI(lc LoadConfig) *TUI {
	// The text input UI element doesn't handle IME input properly.
	// https://github.com/charmbracelet/bubbletea/issues/874
	answer := textinput.New()
	answer.Focus()

	return &TUI{
		LoadConfig: lc,
		answer:     answer,
		help:       help.New(),
		keys: keyMap{
			Quit: key.NewBinding(
				key.WithKeys("esc", "ctrl+c"),
				key.WithHelp("ESC", "quit"),
			),
			Submit: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("ENTER", "submit"),
			),
		},
	}
}

// Init loads flash cards to be reviewed.
func (t *TUI) Init() tea.Cmd {
	flashcards, err := LoadFromCSV(t.LoadConfig)
	if err != nil {
		fmt.Printf("Failed to load flashcards: %v\n", err.Error())
		return tea.Quit
	}
	t.flashcards = flashcards
	return textinput.Blink
}

// Update updates the TUI's state based on user input.
func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, t.keys.Quit):
			return t, tea.Quit
		case key.Matches(msg, t.keys.Submit):
			f := t.flashcards[t.cursor]
			isCorrect := f.Check(t.answer.Value())
			if !t.showExpected {
				if isCorrect {
					t.correctCount++
				} else {
					t.showExpected = true
				}
				t.viewCount++
			}
			if isCorrect {
				t.showExpected = false
				t.answer.Reset()
				if t.cursor < len(t.flashcards)-1 {
					t.cursor++
				} else {
					t.cursor = 0
				}
			}
		}
	}
	t.answer, cmd = t.answer.Update(msg)
	return t, cmd
}

// View returns a string representation of the TUI's current state.
func (t *TUI) View() string {
	f := t.flashcards[t.cursor]

	var correctPerc int
	if t.viewCount > 0 {
		correctPerc = 100 * t.correctCount / t.viewCount
	}

	prompt := QualifiedPrompt(f.Prompt, f.Context)
	if t.showExpected {
		prompt += " > " + f.Answers[0]
	}

	return fmt.Sprintf("Correct: %d/%d (%d%%)\n\n%s\n\n%s\n\n%s",
		t.correctCount,
		t.viewCount,
		correctPerc,
		prompt,
		t.answer.View(),
		t.help.View(t.keys),
	)
}
