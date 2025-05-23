package app

import (
	"encoding/json"
	"fmt"
		"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const batchSize = 10

// TUI is a terminal user interface for reviewing flash cards.
type TUI struct {
	// session is the current review session.
	session *ReviewSession
	// ShowExpected is true if the expected answer should be shown.
	showExpected bool
	// answer is a text input field where the user should enter the answer.
	answer textinput.Model
	// help is the state of the UI element for displaying keybindings.
	help help.Model
	// keys are the keybindings used by the UI.
	keys KeyMap
	// backupPath is the file path where the TUI state will be backed up.
	backupPath string
	// log can be used to write logs to a file.
	log *os.File
}

// NewTUI returns a new TUI.
func NewTUI(lc LoadConfig, backupPath string, log *os.File) (*TUI, error) {
	// The text input UI element doesn't handle IME input properly.
	// https://github.com/charmbracelet/bubbletea/issues/874
	answer := textinput.New()
	answer.Focus()

	session, err := NewReviewSession(lc)
	if err != nil {
		return nil, err
	}

	return &TUI{
		session:    session,
		answer:     answer,
		help:       help.New(),
		keys:       NewKeyMap(),
		backupPath: backupPath,
		log:        log,
	}, nil
}

// LoadTUI loads a backed up TUI state from a file.
func LoadTUI(backupPath string, log *os.File) (*TUI, error) {
	answer := textinput.New()
	answer.Focus()

	session, err := LoadReviewSession(backupPath)
	if err != nil {
		return nil, err
	}

	return &TUI{
		session:    session,
		answer:     answer,
		help:       help.New(),
		keys:       NewKeyMap(),
		backupPath: backupPath,
		log:        log,
	}, nil
}

// Init loads flash cards to be reviewed.
func (t *TUI) Init() tea.Cmd {
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
			ok := t.session.Submit(t.answer.Value(), !t.showExpected)
			if !ok {
				t.showExpected = true
				break
			}

			// Once the user provides the correct answer, we can reset the UI and back up the state.
			t.showExpected = false
			t.answer.Reset()
			err := t.saveToFile()
			if err != nil {
				fmt.Fprintf(t.log, "Couldn't save current state: %v", err) // nolint:errcheck
			}
		}
	}
	t.answer, cmd = t.answer.Update(msg)
	return t, cmd
}

// View returns a string representation of the TUI's current state.
func (t *TUI) View() string {
	f := t.session.Current[0]

	prompt := QualifiedPrompt(f.Prompt, f.Context)
	if t.showExpected {
		prompt += " > " + f.Answers[0]
	}

	output := "Deck counts"
	output += fmt.Sprintf(" · %d", len(t.session.Unreviewed))
	output += fmt.Sprintf(" · %d", len(t.session.Current))
	for _, deck := range t.session.Decks {
		output += fmt.Sprintf(" · %d", len(deck))
	}
	output += fmt.Sprintf(" · Current session: %d/%d (%d%%)\n\n%s\n\n%s\n\n%s\n",
		t.session.correctCount,
		t.session.viewCount,
		percent(t.session.correctCount, t.session.viewCount),
		prompt,
		t.answer.View(),
		t.help.View(t.keys),
	)
	return output
}

func (t *TUI) saveToFile() error {
	b, err := json.MarshalIndent(t.session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(t.backupPath, b, 0600)
}

func percent(numerator, denominator int) int {
	if denominator == 0 {
		return 0
	}
	return 100 * numerator / denominator
}
