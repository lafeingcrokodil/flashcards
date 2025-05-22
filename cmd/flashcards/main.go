package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lafeingcrokodil/flashcards/app"
)

const backupPath = "tmp/backup.json"
const debugPath = "tmp/debug.log"

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

	log, err := tea.LogToFile(debugPath, "")
	if err != nil {
		return err
	}
	defer log.Close() // nolint:errcheck

	var tui *app.TUI
	_, err = os.Stat(backupPath)
	if errors.Is(err, os.ErrNotExist) {
		lc := app.LoadConfig{
			Filepath:      "data/translations.tsv",
			Delimiter:     '\t',
			PromptHeader:  "english",
			ContextHeader: "context",
			AnswerHeader:  "korean",
		}
		tui, err = app.NewTUI(lc, backupPath, log)
	} else {
		tui, err = app.LoadTUI(backupPath, log)
	}
	if err != nil {
		return err
	}

	p := tea.NewProgram(tui)
	_, err = p.Run()
	return err
}
