package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/lafeingcrokodil/flashcards/review"
	"github.com/lafeingcrokodil/flashcards/web"
)

const port = 8080
const backupPath = "tmp/backup.json"

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR\t%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	err := os.MkdirAll(path.Dir(backupPath), os.ModePerm)
	if err != nil {
		return err
	}

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

	server, err := web.New(ctx, r, store)
	if err != nil {
		return err
	}

	return server.Start(port)
}
