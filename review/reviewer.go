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
	source FlashcardMetadataSource
	store  SessionStore
}

// CreateSession creates a new session with all flashcards marked as unreviewed.
func (r *Reviewer) CreateSession(ctx context.Context) (*SessionMetadata, error) {
	sessionID := uuid.NewString()

	flashcardMetadata, err := r.getFlashcardMetadata(ctx)
	if err != nil {
		return nil, err
	}

	sessionMetadata := NewSessionMetadata(sessionID)
	sessionMetadata.UnreviewedCount = len(flashcardMetadata)

	err = r.store.SetSessionMetadata(ctx, sessionID, sessionMetadata)
	if err != nil {
		return nil, err
	}

	err = r.store.SetFlashcards(ctx, sessionID, flashcardMetadata)
	if err != nil {
		return nil, err
	}

	return sessionMetadata, nil
}

// GetSession returns an existing session.
func (r *Reviewer) GetSession(ctx context.Context, sessionID string) (*SessionMetadata, error) {
	return r.store.GetSessionMetadata(ctx, sessionID)
}

// GetFlashcards returns all flashcards.
func (r *Reviewer) GetFlashcards(ctx context.Context, sessionID string) ([]*Flashcard, error) {
	return r.store.GetFlashcards(ctx, sessionID)
}

// SyncFlashcards ensures that the session data is up to date with the flashcard metadata source.
func (r *Reviewer) SyncFlashcards(ctx context.Context, sessionID string) (*SessionMetadata, error) {
	session, err := r.store.GetSessionMetadata(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	flashcardMetadata, err := r.getFlashcardMetadata(ctx)
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

	err = r.store.SetSessionMetadata(ctx, sessionID, updatedSession)
	if err != nil {
		return nil, err
	}

	return updatedSession, nil
}

// NextFlashcard returns the next flashcard to be reviewed.
func (r *Reviewer) NextFlashcard(ctx context.Context, sessionID string) (*Flashcard, error) {
	metadata, err := r.store.GetSessionMetadata(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if metadata.IsNewRound {
		f, err := r.store.NextUnreviewed(ctx, sessionID)
		if !errors.Is(err, ErrNotFound) {
			return f, err
		}
	}

	f, err := r.store.NextReviewed(ctx, sessionID, metadata.Round)
	if !errors.Is(err, ErrNotFound) {
		return f, err
	}

	metadata.Round++
	metadata.IsNewRound = true

	err = r.store.SetSessionMetadata(ctx, sessionID, metadata)
	if err != nil {
		return nil, err
	}

	return r.NextFlashcard(ctx, sessionID)
}

// Submit updates the session state following the review of a flashcard.
func (r *Reviewer) Submit(ctx context.Context, sessionID string, flashcardID int64, submission *Submission) (*SessionMetadata, bool, error) {
	metadata, err := r.store.GetSessionMetadata(ctx, sessionID)
	if err != nil {
		return nil, false, err
	}

	f, err := r.store.GetFlashcard(ctx, sessionID, flashcardID)
	if err != nil {
		return nil, false, err
	}

	previousViewCount := f.Stats.ViewCount
	previousRepetitions := f.Stats.Repetitions

	ok := f.Submit(submission, metadata.Round)
	if !ok {
		return metadata, ok, nil
	}

	err = r.store.SetFlashcardStats(ctx, sessionID, f.Metadata.ID, &f.Stats)
	if err != nil {
		return nil, false, err
	}

	i := proficiencyIndex(f.Stats.Repetitions)
	metadata.ProficiencyCounts[i]++

	if previousViewCount != 0 {
		prev := proficiencyIndex(previousRepetitions)
		metadata.ProficiencyCounts[prev]--
	} else {
		metadata.UnreviewedCount--
	}

	if metadata.IsNewRound {
		metadata.IsNewRound = false
	}

	err = r.store.SetSessionMetadata(ctx, sessionID, metadata)
	if err != nil {
		return nil, false, err
	}

	return metadata, ok, nil
}

func (r *Reviewer) getFlashcardMetadata(ctx context.Context) ([]*FlashcardMetadata, error) {
	metadata, err := r.source.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	metadataByPrompt := make(map[string]*FlashcardMetadata)

	for _, m := range metadata {
		e, ok := metadataByPrompt[m.Prompt]
		if ok && e.Answer != m.Answer {
			return nil, fmt.Errorf("answers %s and %s for prompt %s: %w",
				e.Answer,
				m.Answer,
				m.Prompt,
				ErrAmbiguousAnswers,
			)
		}
		metadataByPrompt[m.Prompt] = m
	}

	return metadata, nil
}

func diff(
	session *SessionMetadata,
	flashcards []*Flashcard,
	metadata []*FlashcardMetadata,
) (updatedSession *SessionMetadata, toBeDeleted []int64, toBeUpserted []*FlashcardMetadata) {
	metadataByID := make(map[int64]*FlashcardMetadata, len(metadata))
	for _, m := range metadata {
		metadataByID[m.ID] = m
	}

	updatedSession = NewSessionMetadata(session.ID)
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

		if f.Metadata != *m {
			fmt.Printf("INFO\tUpdating metadata for ID %d: %v > %v\n", m.ID, f.Metadata, m)
			toBeUpserted = append(toBeUpserted, m)
			updatedSession.UnreviewedCount++
		} else if f.Stats.ViewCount == 0 {
			updatedSession.UnreviewedCount++
		} else {
			i := proficiencyIndex(f.Stats.Repetitions)
			updatedSession.ProficiencyCounts[i]++
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
