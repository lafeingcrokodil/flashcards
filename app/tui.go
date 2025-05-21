package app

import (
	"fmt"

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
	// answer is the user-provided answer.
	answer textinput.Model
}

// NewTUI returns a new TUI.
func NewTUI(lc LoadConfig) *TUI {
	answer := textinput.New()
	answer.Focus()

	return &TUI{
		LoadConfig: lc,
		answer:     answer,
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
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return t, tea.Quit
		case tea.KeyEnter:
			f := t.flashcards[t.cursor]
			if f.Check(t.answer.Value()) {
				t.correctCount++
			}
			t.viewCount++
			t.answer.Reset()
			if t.cursor < len(t.flashcards)-1 {
				t.cursor++
			} else {
				t.cursor = 0
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

	return fmt.Sprintf(`
Correct: %d/%d (%d%%)

%s

%s
	`,
		t.correctCount,
		t.viewCount,
		correctPerc,
		QualifiedPrompt(f.Prompt, f.Context),
		t.answer.View(),
	)
}
