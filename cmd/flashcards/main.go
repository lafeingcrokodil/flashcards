package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lafeingcrokodil/flashcards/app"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		return err
	}

	log, err := tea.LogToFile("tmp/debug.log", "")
	if err != nil {
		return err
	}
	defer log.Close() // nolint:errcheck

	lc := app.LoadConfig{
		Filepath:      "data/translations.tsv",
		Delimiter:     '\t',
		PromptHeader:  "english",
		ContextHeader: "context",
		AnswerHeader:  "korean",
	}

	tui, err := app.NewTUI(lc, log)
	if err != nil {
		return err
	}

	p := tea.NewProgram(tui)
	_, err = p.Run()
	return err
}
