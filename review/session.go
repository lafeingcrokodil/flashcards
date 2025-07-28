package review

import "context"

const numProficiencyLevels = 5

// SessionStore stores the state of a review session.
type SessionStore interface {
	// DeleteFlashcards deletes the specified flashcards.
	DeleteFlashcards(ctx context.Context, sessionID string, ids []int64) error
	// GetFlashcard returns the specified flashcard.
	GetFlashcard(ctx context.Context, sessionID string, flashcardID int64) (*Flashcard, error)
	// GetFlashcards returns all flashcards.
	GetFlashcards(ctx context.Context, sessionID string) ([]*Flashcard, error)
	// SetFlashcards upserts the specified flashcards, clearing any existing stats.
	SetFlashcards(ctx context.Context, sessionID string, metadata []*FlashcardMetadata) error
	// SetFlashcardStats updates a flashcard's stats.
	SetFlashcardStats(ctx context.Context, sessionID string, flashcardID int64, stats *FlashcardStats) error
	// NextReviewed returns a flashcard that is due to be reviewed again.
	NextReviewed(ctx context.Context, sessionID string, round int) (*Flashcard, error)
	// NextUnreviewed returns a flashcard that has never been reviewed before.
	NextUnreviewed(ctx context.Context, sessionID string) (*Flashcard, error)
	// GetSessionMetadata returns the current session metadata.
	GetSessionMetadata(ctx context.Context, sessionID string) (*SessionMetadata, error)
	// SetSessionMetadata updates the session metadata.
	SetSessionMetadata(ctx context.Context, sessionID string, metadata *SessionMetadata) error
}

// SessionMetadata represents review session metadata.
type SessionMetadata struct {
	// ID uniquely identifies a review session.
	ID string `firestore:"id" json:"id"`
	// Round is an incrementing counter that identifies the current round.
	Round int `firestore:"round"`
	// IsNewRound is true if and only if the round has just started.
	IsNewRound bool `firestore:"new"`
	// ProficiencyCounts is the number of flashcards at each proficiency level.
	ProficiencyCounts []int `firestore:"proficiencyCounts" json:"proficiencyCounts"`
	// UnreviewedCount is the number of flashcards that haven't been reviewed yet.
	UnreviewedCount int `firestore:"unreviewedCount" json:"unreviewedCount"`
}

// NewSessionMetadata initializes session metadata for the case where no flashcards have been added yet.
func NewSessionMetadata(sessionID string) *SessionMetadata {
	return &SessionMetadata{
		ID:                sessionID,
		IsNewRound:        true,
		ProficiencyCounts: make([]int, numProficiencyLevels),
	}
}

func proficiencyIndex(repetitions int) int {
	return min(repetitions, numProficiencyLevels-1)
}
