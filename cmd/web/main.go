package main

import (
	"fmt"
	"os"

	"github.com/lafeingcrokodil/flashcards/review"
	"github.com/lafeingcrokodil/flashcards/web"
)

const port = 8080
const backupPath = "tmp/backup.json"

func main() {
	if err := run(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	lc := review.LoadConfig{
		Filepath:      "data/translations.tsv",
		Delimiter:     '\t',
		PromptHeader:  "english",
		ContextHeader: "context",
		AnswerHeader:  "korean",
	}

	s, err := review.NewSession(lc, backupPath)
	if err != nil {
		return err
	}

	server := web.Server{Session: s, BackupPath: backupPath}

	return server.Start(port)
}
