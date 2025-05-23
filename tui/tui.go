package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lafeingcrokodil/flashcards/io"
	"github.com/lafeingcrokodil/flashcards/math"
	"github.com/lafeingcrokodil/flashcards/review"
)

// TUI is a terminal user interface for reviewing flash cards.
type TUI struct {
	// session is the current review session.
	session *review.Session
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

// New returns a new TUI.
func New(lc review.LoadConfig, backupPath string, log *os.File) (*TUI, error) {
	// The text input UI element doesn't handle IME input properly.
	// https://github.com/charmbracelet/bubbletea/issues/874
	answer := textinput.New()
	answer.Focus()

	s, err := review.NewSession(lc, backupPath)
	if err != nil {
		return nil, err
	}

	return &TUI{
		session:    s,
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
			err := io.WriteJSONFile(t.backupPath, t.session)
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
	if t.showExpected {
		expected = expectedStyle.Render("Expected: " + f.Answers[0])
	}

	return fmt.Sprintf("%d%s | %d 👁 · %d ✔ · %d ✖ · %d%%\n\n%s%s\n\n%s\n%s\n\n%s\n",
		len(t.session.Unreviewed),
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
