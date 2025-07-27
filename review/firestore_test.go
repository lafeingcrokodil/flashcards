package review

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/require"
)

func TestFirestoreStore(t *testing.T) {
	ctx := context.Background()

	projectID := os.Getenv("FLASHCARDS_FIRESTORE_PROJECT")
	databaseID := os.Getenv("FLASHCARDS_FIRESTORE_DATABASE")
	collection := os.Getenv("FLASHCARDS_FIRESTORE_COLLECTION")

	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
	require.NoError(t, err)
	defer client.Close() // nolint:errcheck

	_, err = NewFirestoreStore(ctx, client, collection, "")
	require.NoError(t, err)
}
