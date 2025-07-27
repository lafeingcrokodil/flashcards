package review

import (
	"cmp"
	"context"
	"fmt"
	"slices"
)

// MemoryStore stores a review session's state in memory. It's unoptimized and
// mainly intended for use in tests.
type MemoryStore struct {
	metadata   *SessionMetadata
	flashcards []*Flashcard
}

// NewMemoryStore returns a new empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		metadata: &SessionMetadata{
			Round: 0,
			New:   true,
		},
	}
}

// BulkSyncFlashcards aligns the session data with the source of truth for the flashcard metadata.
func (s *MemoryStore) BulkSyncFlashcards(_ context.Context, metadata []*FlashcardMetadata) error {
	var syncedFlashcards []*Flashcard

	metadataByID := make(map[int64]*FlashcardMetadata, len(metadata))
	for _, m := range metadata {
		metadataByID[m.ID] = m
	}

	// Update and clean up existing flashcards.
	for _, e := range s.flashcards {
		m, ok := metadataByID[e.Metadata.ID]
		if !ok {
			fmt.Printf("INFO\tRemoving flashcard with ID %d (%s)\n", e.Metadata.ID, e.Metadata.Answer)
			continue
		}

		if e.Metadata != *m {
			fmt.Printf("INFO\tUpdating metadata for ID %d: %v > %v\n", m.ID, e.Metadata, m)
			e.Metadata = *m
			e.Stats = FlashcardStats{}
		}

		syncedFlashcards = append(syncedFlashcards, e)

		// The flashcard should now be fully synced, so no further processing is required.
		delete(metadataByID, e.Metadata.ID)
	}

	// Add any missing flashcards.
	for _, m := range metadataByID {
		fmt.Printf("INFO\tAdding flashcard with ID %d (%s)\n", m.ID, m.Answer)
		syncedFlashcards = append(syncedFlashcards, &Flashcard{Metadata: *m})
	}

	// Ensure deterministic ordering.
	slices.SortFunc(syncedFlashcards, func(a, b *Flashcard) int {
		return cmp.Compare(a.Metadata.ID, b.Metadata.ID)
	})

	s.flashcards = syncedFlashcards

	return nil
}

// GetFlashcards returns all flashcards.
func (s *MemoryStore) GetFlashcards(ctx context.Context) ([]*Flashcard, error) {
	return s.flashcards, nil
}

// NextReviewed returns a flashcard that is due to be reviewed again.
func (s *MemoryStore) NextReviewed(_ context.Context, round int) (*Flashcard, error) {
	for _, f := range s.flashcards {
		if f.Stats.ViewCount > 0 && f.Stats.NextReview <= round {
			return f, nil
		}
	}
	return nil, nil
}

// NextUnreviewed returns a flashcard that has never been reviewed before.
func (s *MemoryStore) NextUnreviewed(_ context.Context) (*Flashcard, error) {
	for _, f := range s.flashcards {
		if f.Stats.ViewCount == 0 {
			return f, nil
		}
	}
	return nil, nil
}

// GetSessionMetadata returns the current session metadata.
func (s *MemoryStore) GetSessionMetadata(_ context.Context) (*SessionMetadata, error) {
	return s.metadata, nil
}

// UpdateFlashcardStats updates a flashcard's stats.
func (s *MemoryStore) UpdateFlashcardStats(_ context.Context, flashcardID int64, stats *FlashcardStats) error {
	for _, existing := range s.flashcards {
		if existing.Metadata.ID == flashcardID {
			existing.Stats = *stats
		}
	}
	return nil
}

// SetSessionMetadata updates the session metadata.
func (s *MemoryStore) SetSessionMetadata(_ context.Context, metadata *SessionMetadata) error {
	s.metadata = metadata
	return nil
}
