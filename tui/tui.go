package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lafeingcrokodil/flashcards/math"
	"github.com/lafeingcrokodil/flashcards/review"
)

// TUI is a terminal user interface for reviewing flash cards.
type TUI struct {
	// session is the current review session.
	session *review.Session
	// isFirstGuess is true if the expected answer should be shown.
	isFirstGuess bool
	// answer is a text input field where the user should enter the answer.
	answer textinput.Model
	// help is the state of the UI element for displaying keybindings.
	help help.Model
	// keys are the keybindings used by the UI.
	keys KeyMap
	// store is where the TUI state will be backed up.
	store review.SessionStore
	// log can be used to write logs to a file.
	log *os.File
}

// New initializes a new TUI.
func New(ctx context.Context, fr review.FlashcardReader, store review.SessionStore, log *os.File) (*TUI, error) {
	// The text input UI element doesn't handle IME input properly.
	// https://github.com/charmbracelet/bubbletea/issues/874
	answer := textinput.New()
	answer.Focus()

	s, err := review.NewSession(ctx, fr, store)
	if err != nil {
		return nil, err
	}

	return &TUI{
		session:      s,
		isFirstGuess: true,
		answer:       answer,
		help:         help.New(),
		keys:         NewKeyMap(),
		store:        store,
		log:          log,
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
			ok := t.session.Submit(t.answer.Value(), t.isFirstGuess)
			if !ok {
				t.isFirstGuess = false
				break
			}

			// Once the user provides the correct answer, we can reset the UI and back up the state.
			t.isFirstGuess = true
			t.answer.Reset()
			err := t.store.Write(context.Background(), t.session)
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

	var proficiencyCounts string
	for _, count := range t.session.CountByProficiency {
		proficiencyCounts += fmt.Sprintf(" · %d", count)
	}

	var context string
	if f.Context != "" {
		context = contextStyle.Render(" (" + f.Context + ")")
	}

	var expected string
	if !t.isFirstGuess {
		expected = expectedStyle.Render("Expected: " + f.Answer)
	}

	return fmt.Sprintf("%d%s | %d 👁 · %d ✔ · %d ✖ · %d%%\n\n%s%s\n\n%s\n%s\n\n%s\n",
		t.session.UnreviewedCount(),
		proficiencyCounts,
		t.session.ViewCount,
		t.session.CorrectCount,
		t.session.ViewCount-t.session.CorrectCount,
		math.Percent(t.session.CorrectCount, t.session.ViewCount),
		f.Prompt,
		context,
		t.answer.View(),
		expected,
		t.help.View(t.keys),
	)
}
