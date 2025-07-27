package review

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemoryStore(t *testing.T) {
	testSessionStore(t, context.Background(), NewMemoryStore())
}

func testSessionStore(t *testing.T, ctx context.Context, store SessionStore) {
	expectedUnreviewed := &Flashcard{
		Metadata: FlashcardMetadata{ID: 1, Prompt: "What is 1?", Answer: "1"},
	}

	expectedReviewed := &Flashcard{
		Metadata: FlashcardMetadata{ID: 1, Prompt: "What is 1?", Answer: "1"},
		Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 1},
	}

	expectedSessionStats := &SessionStats{Round: 1, New: true}

	expectedFlashcards := []*Flashcard{
		{
			Metadata: FlashcardMetadata{ID: 1, Prompt: "What is 1?", Answer: "1"},
			Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 1},
		},
		{
			Metadata: FlashcardMetadata{ID: 2, Prompt: "What is B?", Answer: "2"},
		},
		{
			Metadata: FlashcardMetadata{ID: 3, Prompt: "What is 3?", Answer: "C"},
		},
		{
			Metadata: FlashcardMetadata{ID: 4, Prompt: "What is 4?", Answer: "4", Context: "ctx"},
		},
		{
			Metadata: FlashcardMetadata{ID: 6, Prompt: "What is 6?", Answer: "6"},
		},
	}

	const initialFlashcardCount = 5

	source := newMemorySource(initialFlashcardCount)

	metadata, err := source.ReadAll(ctx)
	require.NoError(t, err)

	err = store.BulkSyncFlashcards(ctx, metadata)
	require.NoError(t, err)

	firstUnreviewed, err := store.NextUnreviewed(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedUnreviewed, firstUnreviewed)

	for i := 1; i <= initialFlashcardCount; i++ {
		stats := flashcardStats(i)
		err := store.UpdateFlashcardStats(ctx, int64(i), &stats)
		require.NoError(t, err, i)
	}

	err = store.UpdateSessionStats(ctx, expectedSessionStats)
	require.NoError(t, err)

	sessionStats, err := store.SessionStats(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedSessionStats, sessionStats)

	secondUnreviewed, err := store.NextUnreviewed(ctx)
	require.NoError(t, err)
	require.Nil(t, secondUnreviewed)

	firstReviewed, err := store.NextReviewed(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, expectedReviewed, firstReviewed)

	secondReviewed, err := store.NextReviewed(ctx, 6)
	require.NoError(t, err)
	require.Nil(t, secondReviewed)

	err = store.BulkSyncFlashcards(ctx, []*FlashcardMetadata{
		{ID: 1, Prompt: "What is 1?", Answer: "1"},
		{ID: 2, Prompt: "What is B?", Answer: "2"},
		{ID: 3, Prompt: "What is 3?", Answer: "C"},
		{ID: 4, Prompt: "What is 4?", Answer: "4", Context: "ctx"},
		{ID: 6, Prompt: "What is 6?", Answer: "6"},
	})
	require.NoError(t, err)

	flashcards, err := store.GetFlashcards(ctx)
	require.NoError(t, err)

	require.Equal(t, expectedFlashcards, flashcards)
}
