package main

import (
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
	err := os.MkdirAll(path.Dir(backupPath), os.ModePerm)
	if err != nil {
		return err
	}

	lc := review.LoadConfig{
		Filepath:      "data/translations.tsv",
		Delimiter:     '\t',
		PromptHeader:  "english",
		ContextHeader: "context",
		AnswerHeader:  "korean",
	}

	server, err := web.New(lc, backupPath)
	if err != nil {
		return err
	}

	return server.Start(port)
}
