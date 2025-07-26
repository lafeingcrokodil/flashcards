package review

import (
	"context"
)

// FlashcardMetadataSource is the source of truth for flashcard metadata.
type FlashcardMetadataSource interface {
	// ReadAll returns the metadata for all flashcards.
	ReadAll(ctx context.Context) ([]*FlashcardMetadata, error)
}

// SessionStats represents review session stats.
type SessionStats struct {
	// Round is an incrementing counter that identifies the current round.
	Round int `firestore:"round"`
	// New is true if and only if the round has just started.
	New bool `firestore:"new"`
}

// SessionStore stores the state of a review session.
type SessionStore interface {
	// BulkSyncFlashcards aligns the session data with the source of truth for the flashcard metadata.
	BulkSyncFlashcards(ctx context.Context, metadata []*FlashcardMetadata) error
	// NextReviewed returns a flashcard that is due to be reviewed again.
	NextReviewed(ctx context.Context, round int) (*Flashcard, error)
	// NextUnreviewed returns a flashcard that has never been reviewed before.
	NextUnreviewed(ctx context.Context) (*Flashcard, error)
	// SessionStats returns the current session stats.
	SessionStats(ctx context.Context) (*SessionStats, error)
	// UpdateFlashcardStats updates a flashcard's stats.
	UpdateFlashcardStats(ctx context.Context, flashcardID int64, stats *FlashcardStats) error
	// UpdateSessionStats updates the session stats.
	UpdateSessionStats(ctx context.Context, stats *SessionStats) error
}

// Reviewer manages the review of a set of flashcards.
type Reviewer struct {
	store SessionStore
}

// NewReviewer returns a new reviewer. Flashcards are loaded from the source
// and the session state is persisted in the store.
func NewReviewer(ctx context.Context, source FlashcardMetadataSource, store SessionStore) (*Reviewer, error) {
	metadata, err := source.ReadAll(ctx)
	if err != nil {
		return nil, err
	}

	err = store.BulkSyncFlashcards(ctx, metadata)
	if err != nil {
		return nil, err
	}

	return &Reviewer{
		store: store,
	}, nil
}

// Next returns the next flashcard to be reviewed.
func (r *Reviewer) Next(ctx context.Context) (*Flashcard, error) {
	stats, err := r.store.SessionStats(ctx)
	if err != nil {
		return nil, err
	}

	if stats.New {
		f, err := r.store.NextUnreviewed(ctx)
		if f != nil || err != nil {
			return f, err
		}
	}

	f, err := r.store.NextReviewed(ctx, stats.Round)
	if f != nil || err != nil {
		return f, err
	}

	err = r.store.UpdateSessionStats(ctx, &SessionStats{
		Round: stats.Round + 1,
		New:   true,
	})
	if err != nil {
		return nil, err
	}

	return r.Next(ctx)
}

// Submit updates the session state following the review of a flashcard.
func (r *Reviewer) Submit(ctx context.Context, f *Flashcard, correct bool) error {
	stats, err := r.store.SessionStats(ctx)
	if err != nil {
		return err
	}

	if stats.New {
		err = r.store.UpdateSessionStats(ctx, &SessionStats{
			Round: stats.Round,
			New:   false,
		})
		if err != nil {
			return err
		}
	}

	f.Update(correct, stats.Round)

	return r.store.UpdateFlashcardStats(ctx, f.Metadata.ID, &f.Stats)
}
