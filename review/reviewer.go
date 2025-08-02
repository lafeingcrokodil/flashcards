// Package review implements business logic for managing flashcard review sessions.
package review

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var (
	// ErrAmbiguousAnswers is thrown if a flashcard has contradictory answers.
	ErrAmbiguousAnswers = errors.New("answers are ambiguous")
	// ErrNotFound is thrown if the specified data isn't found.
	ErrNotFound = errors.New("not found")
)

// Reviewer manages flashcard review sessions.
type Reviewer struct {
	store SessionStore
}

// NewReviewer returns a new flashcard reviewer.
func NewReviewer(store SessionStore) *Reviewer {
	return &Reviewer{store: store}
}

// CreateSession creates a new session with all flashcards marked as unreviewed.
func (r *Reviewer) CreateSession(ctx context.Context, source FlashcardMetadataSource, numProficiencyLevels int) (*Session, error) {
	sessionID := uuid.NewString()

	flashcardMetadata, err := getFlashcardMetadata(ctx, source)
	if err != nil {
		return nil, err
	}

	session := NewSession(sessionID, numProficiencyLevels)
	session.UnreviewedCount = len(flashcardMetadata)

	err = r.store.SetSession(ctx, sessionID, session)
	if err != nil {
		return nil, err
	}

	err = r.store.SetFlashcards(ctx, sessionID, flashcardMetadata)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSessions returns all existing sessions.
func (r *Reviewer) GetSessions(ctx context.Context) ([]*Session, error) {
	return r.store.GetSessions(ctx)
}

// GetSession returns an existing session.
func (r *Reviewer) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return r.store.GetSession(ctx, sessionID)
}

// GetFlashcards returns all flashcards.
func (r *Reviewer) GetFlashcards(ctx context.Context, sessionID string) ([]*Flashcard, error) {
	return r.store.GetFlashcards(ctx, sessionID)
}

// SyncFlashcards ensures that the session data is up to date with the flashcard metadata source.
func (r *Reviewer) SyncFlashcards(ctx context.Context, sessionID string, source FlashcardMetadataSource) (*Session, error) {
	session, err := r.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	flashcardMetadata, err := getFlashcardMetadata(ctx, source)
	if err != nil {
		return nil, err
	}

	existingFlashcards, err := r.store.GetFlashcards(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	updatedSession, toBeDeleted, toBeUpserted := diff(session, existingFlashcards, flashcardMetadata)

	err = r.store.DeleteFlashcards(ctx, sessionID, toBeDeleted)
	if err != nil {
		return nil, err
	}

	err = r.store.SetFlashcards(ctx, sessionID, toBeUpserted)
	if err != nil {
		return nil, err
	}

	err = r.store.SetSession(ctx, sessionID, updatedSession)
	if err != nil {
		return nil, err
	}

	return updatedSession, nil
}

// NextFlashcard returns the next flashcard to be reviewed.
func (r *Reviewer) NextFlashcard(ctx context.Context, sessionID string) (*Flashcard, error) {
	session, err := r.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if session.IsNewRound {
		f, err := r.store.NextUnreviewed(ctx, sessionID)
		if !errors.Is(err, ErrNotFound) {
			return f, err
		}
	}

	f, err := r.store.NextReviewed(ctx, sessionID, session.Round)
	if !errors.Is(err, ErrNotFound) {
		return f, err
	}

	session.Round++
	session.IsNewRound = true

	err = r.store.SetSession(ctx, sessionID, session)
	if err != nil {
		return nil, err
	}

	return r.NextFlashcard(ctx, sessionID)
}

// Submit updates the session state following the review of a flashcard.
func (r *Reviewer) Submit(ctx context.Context, sessionID string, flashcardID int64, submission *Submission) (*Session, bool, error) {
	session, err := r.store.GetSession(ctx, sessionID)
	if err != nil {
		return nil, false, err
	}

	f, err := r.store.GetFlashcard(ctx, sessionID, flashcardID)
	if err != nil {
		return nil, false, err
	}

	previousViewCount := f.Stats.ViewCount
	previousRepetitions := f.Stats.Repetitions

	ok := f.Submit(submission, session.Round)
	if !ok {
		return session, ok, nil
	}

	err = r.store.SetFlashcardStats(ctx, sessionID, f.Metadata.ID, &f.Stats)
	if err != nil {
		return nil, false, err
	}

	session.IncrementProficiency(f.Stats.Repetitions, 1)

	if previousViewCount != 0 {
		session.IncrementProficiency(previousRepetitions, -1)
	} else {
		session.UnreviewedCount--
	}

	if session.IsNewRound {
		session.IsNewRound = false
	}

	err = r.store.SetSession(ctx, sessionID, session)
	if err != nil {
		return nil, false, err
	}

	return session, ok, nil
}

func getFlashcardMetadata(ctx context.Context, source FlashcardMetadataSource) ([]*FlashcardMetadata, error) {
	metadata, err := source.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	metadataByQualifiedPrompt := make(map[qualifiedPrompt]*FlashcardMetadata)

	for _, m := range metadata {
		e, ok := metadataByQualifiedPrompt[m.qualifiedPrompt()]
		if ok && e.Answer != m.Answer {
			return nil, fmt.Errorf("answers %s and %s for prompt %s: %w",
				e.Answer,
				m.Answer,
				m.Prompt,
				ErrAmbiguousAnswers,
			)
		}
		metadataByQualifiedPrompt[m.qualifiedPrompt()] = m
	}

	return metadata, nil
}

func diff(
	session *Session,
	flashcards []*Flashcard,
	metadata []*FlashcardMetadata,
) (updatedSession *Session, toBeDeleted []int64, toBeUpserted []*FlashcardMetadata) {
	metadataByID := make(map[int64]*FlashcardMetadata, len(metadata))
	for _, m := range metadata {
		metadataByID[m.ID] = m
	}

	updatedSession = NewSession(session.ID, len(session.ProficiencyCounts))
	updatedSession.Round = session.Round
	updatedSession.IsNewRound = session.IsNewRound

	// Update and clean up existing flashcards.
	for _, f := range flashcards {
		m, ok := metadataByID[f.Metadata.ID]
		if !ok {
			fmt.Printf("INFO\tRemoving flashcard with ID %d (%s)\n", f.Metadata.ID, f.Metadata.Answer)
			toBeDeleted = append(toBeDeleted, f.Metadata.ID)
			continue
		}

		switch {
		case f.Metadata != *m:
			fmt.Printf("INFO\tUpdating metadata for ID %d: %v > %v\n", m.ID, f.Metadata, m)
			toBeUpserted = append(toBeUpserted, m)
			updatedSession.UnreviewedCount++
		case f.Stats.ViewCount == 0:
			updatedSession.UnreviewedCount++
		default:
			updatedSession.IncrementProficiency(f.Stats.Repetitions, 1)
		}

		// The flashcard should now be fully synced, so no further processing is required.
		delete(metadataByID, f.Metadata.ID)
	}

	// Add any missing flashcards.
	for _, m := range metadataByID {
		fmt.Printf("INFO\tAdding flashcard with ID %d (%s)\n", m.ID, m.Answer)
		toBeUpserted = append(toBeUpserted, m)
		updatedSession.UnreviewedCount++
	}

	return
}
