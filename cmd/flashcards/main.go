package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lafeingcrokodil/flashcards/app"
)

func main() {
	lc := app.LoadConfig{
		Filepath:      "data/translations.tsv",
		Delimiter:     '\t',
		PromptHeader:  "english",
		ContextHeader: "context",
		AnswerHeader:  "korean",
	}
	p := tea.NewProgram(app.NewTUI(lc))
	if _, err := p.Run(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
