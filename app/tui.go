package app

import (
	"encoding/json"
	"fmt"
	"math"
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
			t.handleSubmit()
		}
	}
	t.answer, cmd = t.answer.Update(msg)
	return t, cmd
}

func (t *TUI) handleSubmit() {
	// Get the current flash card.
	f := t.session.Current[0]

	// Check if the submitted answer is correct.
	isCorrect := f.Check(t.answer.Value())

	// If this is a real guess (not a correction to an incorrect answer),
	// then we need to update the stats and move the card to the appropriate deck.
	if !t.showExpected {
		if isCorrect {
			t.session.correctCount++
			if f.Proficiency < len(t.session.Decks)-1 {
				f.Proficiency++
			}
		} else {
			t.showExpected = true
			f.Proficiency = 0
		}
		f.ViewCount++
		t.session.viewCount++
		t.session.Decks[f.Proficiency] = append(t.session.Decks[f.Proficiency], f)
	}

	// Once the user provides the correct answer, we can reset and select the next flashcard.
	if isCorrect {
		defer func() {
			// Save the current state.
			err := t.saveToFile()
			if err != nil {
				fmt.Printf("Couldn't save current state: %v", err)
			}
		}()

		// First, some resetting.
		t.showExpected = false
		t.answer.Reset()

		// If the current round is still in progress, we can just continue.
		if len(t.session.Current) > 1 {
			t.session.Current = t.session.Current[1:]
			return
		}

		// Otherwise, we can start the next round by collecting flashcards from
		// any decks that are scheduled for review.
		t.session.RoundCount++
		t.session.Current = nil
		for i, deck := range t.session.Decks {
			if t.session.RoundCount%int(math.Pow(2, float64(i))) == 0 {
				var popped []Flashcard
				popped, t.session.Decks[i] = pop(deck, batchSize)
				t.session.Current = append(t.session.Current, popped...)
			}
		}

		// The next round will also always include some unreviewed flashcards, if any remain.
		var popped []Flashcard
		popped, t.session.Unreviewed = pop(t.session.Unreviewed, batchSize)
		t.session.Current = append(t.session.Current, popped...)
	}
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
