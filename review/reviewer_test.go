package review

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewReviewer_Ambiguous(t *testing.T) {
	expectedErr := "ambiguous answers for prompt P1: A1 and A2"

	ctx := context.Background()

	source := &MemorySource{
		metadata: []*FlashcardMetadata{
			{ID: 1, Prompt: "P1", Answer: "A1", Context: "C1"},
			{ID: 2, Prompt: "P1", Answer: "A2", Context: "C1"},
		},
	}

	_, err := NewReviewer(ctx, source, NewMemoryStore())
	require.EqualError(t, err, expectedErr)
}

func TestReviewer_Next(t *testing.T) {
	expectedStates := []struct {
		metadata  *SessionMetadata
		flashcard *Flashcard
	}{
		{
			metadata: &SessionMetadata{Round: 0, New: true},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
			},
		},
		{
			metadata: &SessionMetadata{Round: 0, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
			},
		},
		{
			metadata: &SessionMetadata{Round: 1, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 1},
			},
		},
		{
			metadata: &SessionMetadata{Round: 1, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
			},
		},
		{
			metadata: &SessionMetadata{Round: 2, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 2},
			},
		},
		{
			metadata: &SessionMetadata{Round: 2, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 3},
			},
		},
		{
			metadata: &SessionMetadata{Round: 3, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 3},
			},
		},
		{
			metadata: &SessionMetadata{Round: 3, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 4},
			},
		},
		{
			metadata: &SessionMetadata{Round: 4, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 5},
			},
		},
		// No cards to review in round 6.
		{
			metadata: &SessionMetadata{Round: 5, New: false},
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
		metadata, err := r.store.GetSessionMetadata(ctx)
		require.NoError(t, err, i)
		require.Equal(t, expected.metadata, metadata, i)

		f, err := r.Next(ctx)
		require.NoError(t, err, i)
		require.Equal(t, expected.flashcard, f, i)

		err = r.Submit(ctx, f, true)
		require.NoError(t, err, i)
	}
}
