package review

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReviewer_Next(t *testing.T) {
	expectedStates := []struct {
		stats     *SessionStats
		flashcard *Flashcard
	}{
		{
			stats: &SessionStats{Round: 0, New: true},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
			},
		},
		{
			stats: &SessionStats{Round: 0, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
			},
		},
		{
			stats: &SessionStats{Round: 1, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 1},
			},
		},
		{
			stats: &SessionStats{Round: 1, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
			},
		},
		{
			stats: &SessionStats{Round: 2, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 2},
			},
		},
		{
			stats: &SessionStats{Round: 2, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 3},
			},
		},
		{
			stats: &SessionStats{Round: 3, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 3},
			},
		},
		{
			stats: &SessionStats{Round: 3, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 4},
			},
		},
		{
			stats: &SessionStats{Round: 4, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 5},
			},
		},
		// No cards to review in round 6.
		{
			stats: &SessionStats{Round: 5, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 3, Repetitions: 3, NextReview: 7},
			},
		},
	}

	ctx := context.Background()

	r, err := NewReviewer(ctx, newMemorySource(3), NewMemoryStore())
	require.NoError(t, err)

	for i, expected := range expectedStates {
		stats, err := r.store.SessionStats(ctx)
		require.NoError(t, err, i)
		require.Equal(t, expected.stats, stats, i)

		f, err := r.Next(ctx)
		require.NoError(t, err, i)
		require.Equal(t, expected.flashcard, f, i)

		err = r.Submit(ctx, f, true)
		require.NoError(t, err, i)
	}
}
