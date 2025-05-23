package app

import (
	"encoding/json"
	"fmt"
	"math"
	rand "math/rand/v2"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const batchSize = 10

// TUI is a terminal user interface for reviewing flash cards.
type TUI struct {
	// Current is the list of flashcards currently being reviewed.
	Current []*Flashcard `json:"current"`
	// Unreviewed is a list of flashcards that haven't yet been reviewed.
	Unreviewed []*Flashcard `json:"unreviewed"`
	// Decks are flashcards that have already been reviewed, grouped by accuracy.
	Decks [][]*Flashcard `json:"decks"`
	// roundCount is the number of completed rounds.
	Round int `json:"round"`
	// ViewCount is the number of flashcards that have been reviewed in this session.
	viewCount int
	// CorrectCount is the number of correct answers so far in this session.
	correctCount int
	// ShowExpected is true if the expected answer should be shown.
	showExpected bool
	// answer is the user-provided answer.
	answer textinput.Model
	// help displays keybindings to the user.
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

	flashcards, err := LoadFromCSV(lc)
	if err != nil {
		return nil, err
	}

	rand.Shuffle(len(flashcards), func(i, j int) {
		flashcards[i], flashcards[j] = flashcards[j], flashcards[i]
	})

	current, unreviewed := pop(flashcards, batchSize)

	return &TUI{
		Current:    current,
		Unreviewed: unreviewed,
		Decks: [][]*Flashcard{
			nil, // flashcards requiring the most repetition
			nil,
			nil,
			nil,
			nil, // flashcards requiring the least repetition
		},
		answer:     answer,
		help:       help.New(),
		keys:       NewKeyMap(),
		backupPath: backupPath,
		log:        log,
	}, nil
}

// LoadTUI loads a backed up TUI state from a file.
func LoadTUI(backupPath string, log *os.File) (*TUI, error) {
	b, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, err
	}

	var t TUI
	err = json.Unmarshal(b, &t)
	if err != nil {
		return nil, err
	}

	answer := textinput.New()
	answer.Focus()
	t.answer = answer

	t.help = help.New()
	t.keys = NewKeyMap()
	t.backupPath = backupPath
	t.log = log

	return &t, nil
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
	f := t.Current[0]

	// Check if the submitted answer is correct.
	isCorrect := f.Check(t.answer.Value())

	// If this is a real guess (not a correction to an incorrect answer),
	// then we need to update the stats and move the card to the appropriate deck.
	if !t.showExpected {
		if isCorrect {
			t.correctCount++
			if f.Proficiency < len(t.Decks)-1 {
				f.Proficiency++
			}
		} else {
			t.showExpected = true
			f.Proficiency = 0
		}
		t.viewCount++
		t.Decks[f.Proficiency] = append(t.Decks[f.Proficiency], f)
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
		if len(t.Current) > 1 {
			t.Current = t.Current[1:]
			return
		}

		// Otherwise, we can start the next round by collecting flashcards from
		// any decks that are scheduled for review.
		t.Round++
		t.Current = nil
		for i, deck := range t.Decks {
			if t.Round%int(math.Pow(2, float64(i))) == 0 {
				var popped []*Flashcard
				popped, t.Decks[i] = pop(deck, batchSize)
				t.Current = append(t.Current, popped...)
			}
		}

		// The next round will also always include some unreviewed flashcards, if any remain.
		var popped []*Flashcard
		popped, t.Unreviewed = pop(t.Unreviewed, batchSize)
		t.Current = append(t.Current, popped...)
	}
}

// View returns a string representation of the TUI's current state.
func (t *TUI) View() string {
	f := t.Current[0]

	prompt := QualifiedPrompt(f.Prompt, f.Context)
	if t.showExpected {
		prompt += " > " + f.Answers[0]
	}

	output := "Deck counts"
	output += fmt.Sprintf(" · %d", len(t.Unreviewed))
	output += fmt.Sprintf(" · %d", len(t.Current))
	for _, deck := range t.Decks {
		output += fmt.Sprintf(" · %d", len(deck))
	}
	output += fmt.Sprintf(" · Current session: %d/%d (%d%%)\n\n%s\n\n%s\n\n%s\n",
		t.correctCount,
		t.viewCount,
		percent(t.correctCount, t.viewCount),
		prompt,
		t.answer.View(),
		t.help.View(t.keys),
	)
	return output
}

func (t *TUI) saveToFile() error {
	b, err := json.MarshalIndent(t, "", "  ")
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

func pop[T any](queue []T, numElemsToPop int) ([]T, []T) {
	if len(queue) > numElemsToPop {
		return queue[0:numElemsToPop], queue[numElemsToPop:]
	}
	return queue, nil
}
