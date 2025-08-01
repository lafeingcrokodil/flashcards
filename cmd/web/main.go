// Package main contains code for starting a web server for reviewing flashcards.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/lafeingcrokodil/flashcards/review"
	"github.com/lafeingcrokodil/flashcards/web"
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("ERROR\t%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	portStr := os.Getenv("FLASHCARDS_PORT")
	if portStr == "" {
		portStr = os.Getenv("PORT")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	numProficiencyLevels, err := strconv.Atoi(os.Getenv("FLASHCARDS_PROFICIENCY_LEVELS"))
	if err != nil {
		return err
	}

	source := &review.SheetSource{
		SpreadsheetID: os.Getenv("FLASHCARDS_SHEETS_ID"),
		CellRange:     os.Getenv("FLASHCARDS_SHEETS_CELL_RANGE"),
		IDHeader:      os.Getenv("FLASHCARDS_SHEETS_ID_HEADER"),
		PromptHeader:  os.Getenv("FLASHCARDS_SHEETS_PROMPT_HEADER"),
		ContextHeader: os.Getenv("FLASHCARDS_SHEETS_CONTEXT_HEADER"),
		AnswerHeader:  os.Getenv("FLASHCARDS_SHEETS_ANSWER_HEADER"),
	}

	projectID := os.Getenv("FLASHCARDS_FIRESTORE_PROJECT")
	collection := os.Getenv("FLASHCARDS_FIRESTORE_COLLECTION")

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close() //nolint:errcheck

	store := review.NewFirestoreStore(client, collection)

	server, err := web.New(source, store, numProficiencyLevels)
	if err != nil {
		return err
	}

	return server.Start(port)
}
