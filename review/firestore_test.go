package review

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFirestoreStore(t *testing.T) {
	sessionID := uuid.NewString()

	expectedSession := &SessionMetadata{
		ID:                sessionID,
		Round:             2,
		IsNewRound:        true,
		ProficiencyCounts: []int{1, 0, 1, 0, 0},
		UnreviewedCount:   2,
	}

	expectedMetadata := []*FlashcardMetadata{
		{ID: 1, Prompt: "P1", Answer: "A1", Context: "C1"},
		{ID: 2, Prompt: "P2", Answer: "A2"},
		{ID: 3, Prompt: "P3", Answer: "A3"},
		{ID: 4, Prompt: "P4", Answer: "A4"},
	}

	expectedFlashcardStats := []*FlashcardStats{
		{ViewCount: 2, Repetitions: 2, NextReview: 3},
		{ViewCount: 1, Repetitions: 0, NextReview: 2},
	}

	expectedFirstFlashcard := &Flashcard{
		Metadata: *expectedMetadata[0],
		Stats:    *expectedFlashcardStats[0],
	}

	expectedUnreviewedFlashcard := &Flashcard{
		Metadata: *expectedMetadata[2],
	}

	expectedReviewedFlashcard := &Flashcard{
		Metadata: *expectedMetadata[1],
		Stats:    *expectedFlashcardStats[1],
	}

	expectedNotFoundError := "not found"

	expectedUpdatedMetadata := &FlashcardMetadata{ID: 1, Prompt: "P1", Answer: "B1", Context: "C1"}

	expectedFinalFlashcards := []*Flashcard{
		{Metadata: *expectedUpdatedMetadata},
		{Metadata: *expectedMetadata[1], Stats: *expectedFlashcardStats[1]},
	}

	ctx := context.Background()

	projectID := os.Getenv("FLASHCARDS_FIRESTORE_PROJECT")
	collection := os.Getenv("FLASHCARDS_FIRESTORE_COLLECTION")

	client, err := firestore.NewClient(ctx, projectID)
	require.NoError(t, err)
	defer client.Close() //nolint:errcheck

	store := NewFirestoreStore(client, collection)

	err = store.SetSessionMetadata(ctx, sessionID, expectedSession)
	require.NoError(t, err)

	session, err := store.GetSessionMetadata(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, expectedSession, session)

	err = store.SetFlashcards(ctx, sessionID, expectedMetadata)
	require.NoError(t, err)

	for i, stats := range expectedFlashcardStats {
		err = store.SetFlashcardStats(ctx, sessionID, int64(i+1), stats)
		require.NoError(t, err)
	}

	f, err := store.GetFlashcard(ctx, sessionID, 1)
	require.NoError(t, err)
	require.Equal(t, expectedFirstFlashcard, f)

	unreviewed, err := store.NextUnreviewed(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, expectedUnreviewedFlashcard, unreviewed)

	reviewed, err := store.NextReviewed(ctx, sessionID, expectedSession.Round)
	require.NoError(t, err)
	require.Equal(t, expectedReviewedFlashcard, reviewed)

	_, err = store.NextReviewed(ctx, sessionID, 1)
	require.EqualError(t, err, expectedNotFoundError)

	err = store.SetFlashcards(ctx, sessionID, []*FlashcardMetadata{expectedUpdatedMetadata})
	require.NoError(t, err)

	err = store.DeleteFlashcards(ctx, sessionID, []int64{3, 4})
	require.NoError(t, err)

	flashcards, err := store.GetFlashcards(ctx, sessionID)
	require.NoError(t, err)
	require.Equal(t, expectedFinalFlashcards, flashcards)
}
