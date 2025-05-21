package main

import (
	"fmt"
	"os"

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
	fcs, err := app.LoadFromCSV(lc)
	if err != nil {
		fmt.Printf("Failed to load CSV file: %v\n", err)
		os.Exit(1)
	}

	err = app.Review(fcs)
	if err != nil {
		fmt.Printf("Failed to review flashcards: %v\n", err)
		os.Exit(1)
	}
}
