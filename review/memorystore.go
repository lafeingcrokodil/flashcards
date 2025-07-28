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
	metadata   map[string]*SessionMetadata
	flashcards map[string][]*Flashcard
}

// NewMemoryStore returns a new empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		metadata:   make(map[string]*SessionMetadata),
		flashcards: make(map[string][]*Flashcard),
	}
}

// DeleteFlashcards deletes the specified flashcards.
func (s *MemoryStore) DeleteFlashcards(_ context.Context, sessionID string, ids []int64) error {
	flashcards, ok := s.flashcards[sessionID]
	if !ok {
		return fmt.Errorf("flashcards for session %s: %w", sessionID, ErrNotFound)
	}

	var filteredFlashcards []*Flashcard
	for _, f := range flashcards {
		if !slices.Contains(ids, f.Metadata.ID) {
			filteredFlashcards = append(filteredFlashcards, f)
		}
	}

	s.flashcards[sessionID] = filteredFlashcards

	return nil
}

// GetFlashcard returns the specified flashcard.
func (s *MemoryStore) GetFlashcard(_ context.Context, sessionID string, flashcardID int64) (*Flashcard, error) {
	flashcards, ok := s.flashcards[sessionID]
	if !ok {
		return nil, fmt.Errorf("flashcards for session %s: %w", sessionID, ErrNotFound)
	}

	for _, f := range flashcards {
		if f.Metadata.ID == flashcardID {
			return f, nil
		}
	}

	return nil, fmt.Errorf("flashcard %d for session %s: %w", flashcardID, sessionID, ErrNotFound)
}

// GetFlashcards returns all flashcards.
func (s *MemoryStore) GetFlashcards(_ context.Context, sessionID string) ([]*Flashcard, error) {
	flashcards, ok := s.flashcards[sessionID]
	if !ok {
		return nil, fmt.Errorf("flashcards for session %s: %w", sessionID, ErrNotFound)
	}
	return flashcards, nil
}

// SetFlashcards upserts the specified flashcards, clearing any existing stats.
func (s *MemoryStore) SetFlashcards(_ context.Context, sessionID string, metadata []*FlashcardMetadata) error {
	existingFlashcards, ok := s.flashcards[sessionID]
	if !ok {
		existingFlashcards = []*Flashcard{}
	}

	existingFlashcardsByID := make(map[int64]*Flashcard, len(existingFlashcards))
	for _, e := range existingFlashcards {
		existingFlashcardsByID[e.Metadata.ID] = e
	}

	var toBeAppended []*Flashcard

	for _, m := range metadata {
		e, ok := existingFlashcardsByID[m.ID]
		if ok {
			e.Metadata = *m
			e.Stats = FlashcardStats{}
		} else {
			toBeAppended = append(toBeAppended, &Flashcard{Metadata: *m})
		}
	}

	// Ensure deterministic ordering.
	slices.SortFunc(toBeAppended, func(a, b *Flashcard) int {
		return cmp.Compare(a.Metadata.ID, b.Metadata.ID)
	})

	s.flashcards[sessionID] = append(s.flashcards[sessionID], toBeAppended...)

	return nil
}

// SetFlashcardStats updates a flashcard's stats.
func (s *MemoryStore) SetFlashcardStats(_ context.Context, sessionID string, flashcardID int64, stats *FlashcardStats) error {
	flashcards, ok := s.flashcards[sessionID]
	if !ok {
		return fmt.Errorf("flashcards for session %s: %w", sessionID, ErrNotFound)
	}

	var found bool
	for _, f := range flashcards {
		if f.Metadata.ID == flashcardID {
			found = true
			f.Stats = *stats
		}
	}

	if !found {
		return fmt.Errorf("flashcard %d for session %s: %w", flashcardID, sessionID, ErrNotFound)
	}

	return nil
}

// NextReviewed returns a flashcard that is due to be reviewed again.
func (s *MemoryStore) NextReviewed(_ context.Context, sessionID string, round int) (*Flashcard, error) {
	flashcards, ok := s.flashcards[sessionID]
	if !ok {
		return nil, fmt.Errorf("flashcards for session %s: %w", sessionID, ErrNotFound)
	}

	for _, f := range flashcards {
		if f.Stats.ViewCount > 0 && f.Stats.NextReview <= round {
			return f, nil
		}
	}
	return nil, nil
}

// NextUnreviewed returns a flashcard that has never been reviewed before.
func (s *MemoryStore) NextUnreviewed(_ context.Context, sessionID string) (*Flashcard, error) {
	flashcards, ok := s.flashcards[sessionID]
	if !ok {
		return nil, fmt.Errorf("flashcards for session %s: %w", sessionID, ErrNotFound)
	}

	for _, f := range flashcards {
		if f.Stats.ViewCount == 0 {
			return f, nil
		}
	}

	return nil, nil
}

// GetSessionMetadata returns the current session metadata.
func (s *MemoryStore) GetSessionMetadata(_ context.Context, sessionID string) (*SessionMetadata, error) {
	metadata, ok := s.metadata[sessionID]
	if !ok {
		return nil, fmt.Errorf("metadata for session %s: %w", sessionID, ErrNotFound)
	}
	return metadata, nil
}

// SetSessionMetadata updates the session metadata.
func (s *MemoryStore) SetSessionMetadata(_ context.Context, sessionID string, metadata *SessionMetadata) error {
	s.metadata[sessionID] = metadata
	return nil
}
