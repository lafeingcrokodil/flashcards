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

	projectID := os.Getenv("FLASHCARDS_FIRESTORE_PROJECT")
	collection := os.Getenv("FLASHCARDS_FIRESTORE_COLLECTION")

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close() //nolint:errcheck

	store := review.NewFirestoreStore(client, collection)

	server, err := web.New(store, numProficiencyLevels)
	if err != nil {
		return err
	}

	return server.Start(port)
}
