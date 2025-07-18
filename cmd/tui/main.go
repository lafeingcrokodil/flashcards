package main

import (
	"context"
	"fmt"
	"os"
	"path"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lafeingcrokodil/flashcards/review"
	"github.com/lafeingcrokodil/flashcards/tui"
)

const backupPath = "tmp/backup.json"
const debugPath = "tmp/debug.log"

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR\t%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	filepaths := []string{backupPath, debugPath}
	for _, p := range filepaths {
		err := os.MkdirAll(path.Dir(p), os.ModePerm)
		if err != nil {
			return err
		}
	}

	log, err := tea.LogToFile(debugPath, "")
	if err != nil {
		return err
	}
	defer log.Close() // nolint:errcheck

	r := &review.SheetReader{
		SpreadsheetID: "17P6QomOB46SemEFlUhcyzB8fycqltQBjfC4ELzOEy6s",
		CellRange:     "Korean!A:F",
		IDHeader:      "id",
		PromptHeader:  "english",
		ContextHeader: "context",
		AnswerHeader:  "korean",
	}

	store := &review.SheetStore{
		SpreadsheetID: "17P6QomOB46SemEFlUhcyzB8fycqltQBjfC4ELzOEy6s",
		CellRange:     "Session!A:H",
	}

	ui, err := tui.New(ctx, r, store, log)
	if err != nil {
		return err
	}

	p := tea.NewProgram(ui)
	_, err = p.Run()
	return err
}
