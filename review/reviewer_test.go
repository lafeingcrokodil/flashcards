package review

import (
	"context"
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

	memSource := memorySource(3)
	memStore := NewMemoryStore()

	fsStore, closeFunc, err := firestoreStore(ctx)
	if !assert.NoError(t, err) {
		return
	}
	defer closeFunc() // nolint:errcheck

	stores := map[string]SessionStore{
		"memory":    memStore,
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

func TestSessionStore_BulkSyncFlashcards(t *testing.T) {
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

	ctx := context.Background()

	memSource := memorySource(5)

	metadata, err := memSource.ReadAll(ctx)
	if !assert.NoError(t, err) {
		return
	}

	memStore := NewMemoryStore()

	fsStore, closeFunc, err := firestoreStore(ctx)
	if !assert.NoError(t, err) {
		return
	}
	defer closeFunc() // nolint:errcheck

	stores := map[string]SessionStore{
		"memory":    memStore,
		"firestore": fsStore,
	}

	for i, store := range stores {
		err := store.BulkSyncFlashcards(ctx, metadata)
		if !assert.NoError(t, err, i) {
			return
		}

		for j := 1; j < 5; j++ {
			stats := flashcardStats(j)
			err := store.UpdateFlashcardStats(ctx, int64(j), &stats)
			if !assert.NoError(t, err, i, j) {
				return
			}
		}

		err = store.BulkSyncFlashcards(ctx, []*FlashcardMetadata{
			{ID: 1, Prompt: "What is 1?", Answer: "1"},
			{ID: 2, Prompt: "What is B?", Answer: "2"},
			{ID: 3, Prompt: "What is 3?", Answer: "C"},
			{ID: 4, Prompt: "What is 4?", Answer: "4", Context: "ctx"},
			{ID: 6, Prompt: "What is 6?", Answer: "6"},
		})
		if !assert.NoError(t, err, i) {
			return
		}

		flashcards, err := store.GetFlashcards(ctx)
		if !assert.NoError(t, err, i) {
			return
		}

		assert.Equal(t, expectedFlashcards, flashcards)
	}
}

func firestoreStore(ctx context.Context) (*FireStore, func() error, error) {
	projectID := os.Getenv("FLASHCARDS_FIRESTORE_PROJECT")
	databaseID := os.Getenv("FLASHCARDS_FIRESTORE_DATABASE")
	collection := os.Getenv("FLASHCARDS_FIRESTORE_COLLECTION")

	client, err := firestore.NewClientWithDatabase(ctx, projectID, databaseID)
	if err != nil {
		return nil, nil, err
	}
	closeFunc := client.Close

	s, err := NewFireStore(ctx, client, collection, "")
	if err != nil {
		return nil, closeFunc, err
	}

	return s, closeFunc, nil
}

func memorySource(maxID int) *MemorySource {
	memSource := &MemorySource{}

	for i := 1; i <= maxID; i++ {
		metadata := flashcardMetadata(i)
		memSource.metadata = append(memSource.metadata, &metadata)
	}

	return memSource
}
