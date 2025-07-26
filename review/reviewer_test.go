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
				Stats: FlashcardStats{
					ViewCount:   1,
					Repetitions: 1,
					NextReview:  1,
				},
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
				Stats: FlashcardStats{
					ViewCount:   1,
					Repetitions: 1,
					NextReview:  2,
				},
			},
		},
		{
			stats: &SessionStats{Round: 2, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats: FlashcardStats{
					ViewCount:   2,
					Repetitions: 2,
					NextReview:  3,
				},
			},
		},
		{
			stats: &SessionStats{Round: 3, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats: FlashcardStats{
					ViewCount:   1,
					Repetitions: 1,
					NextReview:  3,
				},
			},
		},
		{
			stats: &SessionStats{Round: 3, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats: FlashcardStats{
					ViewCount:   2,
					Repetitions: 2,
					NextReview:  4,
				},
			},
		},
		{
			stats: &SessionStats{Round: 4, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats: FlashcardStats{
					ViewCount:   2,
					Repetitions: 2,
					NextReview:  5,
				},
			},
		},
		// No cards to review in round 6.
		{
			stats: &SessionStats{Round: 5, New: false},
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats: FlashcardStats{
					ViewCount:   3,
					Repetitions: 3,
					NextReview:  7,
				},
			},
		},
	}

	ctx := context.Background()

	memSource := &MemorySource{}

	for i := 1; i <= 3; i++ {
		metadata := flashcardMetadata(int64(i))
		memSource.metadata = append(memSource.metadata, &metadata)
	}

	projectID := os.Getenv("FLASHCARDS_FIRESTORE_PROJECT")
	databaseID := os.Getenv("FLASHCARDS_FIRESTORE_DATABASE")
	collection := os.Getenv("FLASHCARDS_FIRESTORE_COLLECTION")

	fsClient, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
	if !assert.NoError(t, err) {
		return
	}
	defer fsClient.Close() // nolint:errcheck

	fsStore, err := NewFireStore(ctx, fsClient, collection, "")
	if !assert.NoError(t, err) {
		return
	}

	stores := map[string]SessionStore{
		"memory":    NewMemoryStore(),
		"firestore": fsStore,
	}

	for i, store := range stores {
		r, err := NewReviewer(ctx, memSource, store)
		if !assert.NoError(t, err, i) {
			continue
		}

		for j, expected := range expectedStates {
			stats, err := r.store.SessionStats(ctx)
			if !assert.NoError(t, err, i, j) {
				return
			}
			if !assert.Equal(t, expected.stats, stats, i, j) {
				return
			}
			f, err := r.Next(ctx)
			if !assert.NoError(t, err, i, j) {
				return
			}
			if !assert.Equal(t, expected.flashcard, f, i, j) {
				return
			}
			err = r.Submit(ctx, f, true)
			if !assert.NoError(t, err, i, j) {
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
