package review

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
	flashcardMetadata, err := r.getFlashcardMetadata(ctx)
	if err != nil {
		return nil, err
	}

	metadataByID := make(map[int64]*FlashcardMetadata, len(flashcardMetadata))
	for _, m := range flashcardMetadata {
		metadataByID[m.ID] = m
	}

	existingFlashcards, err := r.store.GetFlashcards(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	sessionMetadata, err := r.store.GetSessionMetadata(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	var toBeDeleted []int64
	var toBeUpserted []*FlashcardMetadata

	updatedSessionMetadata := NewSessionMetadata(sessionID)
	updatedSessionMetadata.Round = sessionMetadata.Round
	updatedSessionMetadata.IsNewRound = sessionMetadata.IsNewRound

	// Update and clean up existing flashcards.
	for _, e := range existingFlashcards {
		m, ok := metadataByID[e.Metadata.ID]
		if !ok {
			fmt.Printf("INFO\tRemoving flashcard with ID %d (%s)\n", e.Metadata.ID, e.Metadata.Answer)
			toBeDeleted = append(toBeDeleted, e.Metadata.ID)
			continue
		}

		if e.Metadata != *m {
			fmt.Printf("INFO\tUpdating metadata for ID %d: %v > %v\n", m.ID, e.Metadata, m)
			toBeUpserted = append(toBeUpserted, m)
			updatedSessionMetadata.UnreviewedCount++
		} else if e.Stats.ViewCount == 0 {
			updatedSessionMetadata.UnreviewedCount++
		} else {
			i := proficiencyIndex(e.Stats.Repetitions)
			updatedSessionMetadata.ProficiencyCounts[i]++
		}

		// The flashcard should now be fully synced, so no further processing is required.
		delete(metadataByID, e.Metadata.ID)
	}

	// Add any missing flashcards.
	for _, m := range metadataByID {
		fmt.Printf("INFO\tAdding flashcard with ID %d (%s)\n", m.ID, m.Answer)
		toBeUpserted = append(toBeUpserted, m)
		updatedSessionMetadata.UnreviewedCount++
	}

	err = r.store.DeleteFlashcards(ctx, sessionID, toBeDeleted)
	if err != nil {
		return nil, err
	}

	err = r.store.SetFlashcards(ctx, sessionID, toBeUpserted)
	if err != nil {
		return nil, err
	}

	err = r.store.SetSessionMetadata(ctx, sessionID, updatedSessionMetadata)
	if err != nil {
		return nil, err
	}

	return updatedSessionMetadata, nil
}

// NextFlashcard returns the next flashcard to be reviewed.
func (r *Reviewer) NextFlashcard(ctx context.Context, sessionID string) (*Flashcard, error) {
	metadata, err := r.store.GetSessionMetadata(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if metadata.IsNewRound {
		f, err := r.store.NextUnreviewed(ctx, sessionID)
		if f != nil || err != nil {
			return f, err
		}
	}

	f, err := r.store.NextReviewed(ctx, sessionID, metadata.Round)
	if f != nil || err != nil {
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
			return nil, fmt.Errorf("ambiguous answers for prompt %s: %s and %s",
				m.Prompt,
				e.Answer,
				m.Answer,
			)
		}
		metadataByPrompt[m.Prompt] = m
	}

	return metadata, nil
}
