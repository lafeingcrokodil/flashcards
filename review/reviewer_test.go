package review

import (
	"context"
	"fmt"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
)

func TestReviewer(t *testing.T) {
	expectedFlashcards := []*Flashcard{
		// Round 0
		{
			Metadata: flashcardMetadata(1),
		},
		// Round 1
		{
			Metadata: flashcardMetadata(2),
		},
		{
			Metadata: flashcardMetadata(1),
			Stats: FlashcardStats{
				ViewCount:   1,
				Repetitions: 1,
				NextReview:  1,
			},
		},
		// Round 2
		{
			Metadata: flashcardMetadata(3),
		},
		{
			Metadata: flashcardMetadata(2),
			Stats: FlashcardStats{
				ViewCount:   1,
				Repetitions: 1,
				NextReview:  2,
			},
		},
		// Round 3
		{
			Metadata: flashcardMetadata(1),
			Stats: FlashcardStats{
				ViewCount:   2,
				Repetitions: 2,
				NextReview:  3,
			},
		},
		{
			Metadata: flashcardMetadata(3),
			Stats: FlashcardStats{
				ViewCount:   1,
				Repetitions: 1,
				NextReview:  3,
			},
		},
		// Round 4
		{
			Metadata: flashcardMetadata(2),
			Stats: FlashcardStats{
				ViewCount:   2,
				Repetitions: 2,
				NextReview:  4,
			},
		},
		// Round 5
		{
			Metadata: flashcardMetadata(3),
			Stats: FlashcardStats{
				ViewCount:   2,
				Repetitions: 2,
				NextReview:  5,
			},
		},
		// Round 6 - no cards to review
		// Round 7
		{
			Metadata: flashcardMetadata(1),
			Stats: FlashcardStats{
				ViewCount:   3,
				Repetitions: 3,
				NextReview:  7,
			},
		},
	}

	ctx := context.Background()

	projectID := os.Getenv("FLASHCARDS_FIRESTORE_PROJECT")
	databaseID := os.Getenv("FLASHCARDS_FIRESTORE_DATABASE")
	collection := os.Getenv("FLASHCARDS_FIRESTORE_COLLECTION")

	fsClient, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
	if !assert.NoError(t, err) {
		return
	}
	defer fsClient.Close()

	stores := []FlashcardStore{
		&MemoryStore{},
		&FireStore{
			client:     fsClient,
			collection: collection,
		},
	}

	for _, store := range stores {
		r := NewReviewer(store)

		for i := 1; i <= 3; i++ {
			f := &Flashcard{Metadata: flashcardMetadata(int64(i))}
			err := r.Upsert(ctx, f)
			if !assert.NoError(t, err, f.Metadata.ID) {
				return
			}
		}

		for i, expected := range expectedFlashcards {
			f, err := r.Next(ctx)
			if !assert.NoError(t, err, i) {
				return
			}
			if !assert.Equal(t, expected, f, i) {
				return
			}
			err = r.Submit(ctx, f, true)
			if !assert.NoError(t, err, i) {
				return
			}
		}
	}
}

func flashcardMetadata(id int64) FlashcardMetadata {
	return FlashcardMetadata{
		ID:     id,
		Prompt: fmt.Sprintf("What is %d?", id),
		Answer: fmt.Sprintf("%d", id),
	}
}
