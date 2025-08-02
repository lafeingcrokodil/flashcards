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
	session    map[string]*Session
	flashcards map[string][]*Flashcard
}

// NewMemoryStore returns a new empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		session:    make(map[string]*Session),
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
	return nil, ErrNotFound
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

	return nil, ErrNotFound
}

// GetSession returns the current session metadata.
func (s *MemoryStore) GetSession(_ context.Context, sessionID string) (*Session, error) {
	session, ok := s.session[sessionID]
	if !ok {
		return nil, fmt.Errorf("metadata for session %s: %w", sessionID, ErrNotFound)
	}
	return session, nil
}

// GetSessions returns the metadata for all existing sessions.
func (s *MemoryStore) GetSessions(_ context.Context) ([]*Session, error) {
	sessions := make([]*Session, 0, len(s.session))
	for _, sess := range s.session {
		sessions = append(sessions, sess)
	}
	return sessions, nil
}

// SetSession updates the session metadata.
func (s *MemoryStore) SetSession(_ context.Context, sessionID string, session *Session) error {
	s.session[sessionID] = session
	return nil
}
