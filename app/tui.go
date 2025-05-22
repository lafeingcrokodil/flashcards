package app

import (
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
	// current is the list of flashcards currently being reviewed.
	current []*Flashcard
	// unreviewed is a list of flashcards that haven't yet been reviewed.
	unreviewed []*Flashcard
	// decks are flashcards that have already been reviewed, grouped by accuracy.
	decks [][]*Flashcard
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
	// log can be used to write logs to a file.
	log *os.File
}

// NewTUI returns a new TUI.
func NewTUI(lc LoadConfig, log *os.File) (*TUI, error) {
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
		current:    current,
		unreviewed: unreviewed,
		decks: [][]*Flashcard{
			nil, // flashcards requiring the most repetition
			nil,
			nil,
			nil,
			nil, // flashcards requiring the least repetition
		},
		answer: answer,
		help:   help.New(),
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
		log: log,
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
	f := t.current[0]

	// Check if the submitted answer is correct.
	isCorrect := f.Check(t.answer.Value())

	// If this is a real guess (not a correction to an incorrect answer),
	// then we need to update the stats and move the card to the appropriate deck.
	if !t.showExpected {
		if isCorrect {
			t.correctCount++
			if f.deckIndex < len(t.decks)-1 {
				f.deckIndex++
			}
		} else {
			t.showExpected = true
			f.deckIndex = 0
		}
		t.viewCount++
		t.decks[f.deckIndex] = append(t.decks[f.deckIndex], f)
	}

	// Once the user provides the correct answer, we can reset and select the next flashcard.
	if isCorrect {
		// First, some resetting.
		t.showExpected = false
		t.answer.Reset()

		// If the current round is still in progress, we can just continue.
		if len(t.current) > 1 {
			t.current = t.current[1:]
			return
		}

		// Otherwise, we can start the next round by collecting flashcards from
		// any decks that are scheduled for review.
		t.current = nil
		round := t.viewCount / 10
		for i, deck := range t.decks {
			if round%int(math.Pow(2, float64(i))) == 0 {
				var popped []*Flashcard
				popped, t.decks[i] = pop(deck, batchSize)
				t.current = append(t.current, popped...)
			}
		}

		// The next round will also always include some unreviewed flashcards, if any remain.
		var popped []*Flashcard
		popped, t.unreviewed = pop(t.unreviewed, batchSize)
		t.current = append(t.current, popped...)
	}
}

// View returns a string representation of the TUI's current state.
func (t *TUI) View() string {
	f := t.current[0]

	prompt := QualifiedPrompt(f.Prompt, f.Context)
	if t.showExpected {
		prompt += " > " + f.Answers[0]
	}

	output := fmt.Sprintf("Correct: %d/%d (%d%%)\n\n%s\n\n%s\n\n%s\n",
		t.correctCount,
		t.viewCount,
		percent(t.correctCount, t.viewCount),
		prompt,
		t.answer.View(),
		t.help.View(t.keys),
	)

	output += fmt.Sprintf("\nCurrent: %s", prompts(t.current))
	output += fmt.Sprintf("\nUnreviewed: %s", prompts(t.unreviewed))
	for i, deck := range t.decks {
		output += fmt.Sprintf("\n%d: %s", i, prompts(deck))
	}

	return output
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

func prompts(fs []*Flashcard) []string {
	var prompts []string
	for _, f := range fs {
		prompts = append(prompts, QualifiedPrompt(f.Prompt, f.Context))
	}
	return prompts
}
