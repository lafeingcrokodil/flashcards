package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lafeingcrokodil/flashcards/app"
	"github.com/lafeingcrokodil/flashcards/tui"
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

	var ui *tui.TUI
	_, err = os.Stat(backupPath)
	if errors.Is(err, os.ErrNotExist) {
		lc := app.LoadConfig{
			Filepath:      "data/translations.tsv",
			Delimiter:     '\t',
			PromptHeader:  "english",
			ContextHeader: "context",
			AnswerHeader:  "korean",
		}
		ui, err = tui.New(lc, backupPath, log)
	} else {
		ui, err = tui.Load(backupPath, log)
	}
	if err != nil {
		return err
	}

	p := tea.NewProgram(ui)
	_, err = p.Run()
	return err
}
