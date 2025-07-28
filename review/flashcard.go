package review

import (
	"context"
	"math"
)

const spacedRepetitionFactor = 2

// FlashcardMetadataSource is the source of truth for flashcard metadata.
type FlashcardMetadataSource interface {
	// GetAll returns the metadata for all flashcards.
	GetAll(ctx context.Context) ([]*FlashcardMetadata, error)
}

// Flashcard represents the state of a flashcard.
type Flashcard struct {
	// Metadata stores immutable data like the prompt and answer.
	Metadata FlashcardMetadata `firestore:"metadata"`
	// Stats stores mutable data like the view count.
	Stats FlashcardStats `firestore:"stats"`
}

// FlashcardMetadata stores immutable flashcard data like the prompt.
type FlashcardMetadata struct {
	// ID uniquely identifies the flashcard.
	ID int64 `firestore:"id" json:"id"`
	// Prompt is the text to be shown to the user.
	Prompt string `firestore:"prompt" json:"prompt"`
	// Context helps narrow down possible answers.
	Context string `firestore:"context,omitempty" json:"context,omitempty"`
	// Answer is the accepted answer.
	Answer string `firestore:"answer" json:"answer"`
}

// FlashcardStats stores mutable flashcard data like the view count.
type FlashcardStats struct {
	// ViewCount is the number of times the flashcard has been reviewed.
	ViewCount int `firestore:"viewCount"`
	// Repetitions is the number of successful reviews in a row.
	Repetitions int `firestore:"repetitions,omitempty"`
	// NextReview is the round in which the card is due to be reviewed next.
	NextReview int `firestore:"nextReview,omitempty"`
}

// Submission represents a user's answer to a flashcard prompt.
type Submission struct {
	// Answer is the submitted answer.
	Answer string `json:"answer"`
	// IsFirstGuess is true if and only if this is the user's first guess.
	IsFirstGuess bool `firestore:"isFirstGuess"`
}

type qualifiedPrompt struct {
	prompt  string
	context string
}

// Submit updates the flashcard's stats after being reviewed.
// Returns true if and only if the answer is correct.
func (f *Flashcard) Submit(submission *Submission, round int) bool {
	if submission.Answer != f.Metadata.Answer {
		return false
	}

	f.Stats.ViewCount++

	if submission.IsFirstGuess {
		f.Stats.NextReview = round + interval(f.Stats.Repetitions)
		f.Stats.Repetitions++
	} else {
		f.Stats.NextReview = round + 1
		f.Stats.Repetitions = 0
	}

	return true
}

func (m *FlashcardMetadata) qualifiedPrompt() qualifiedPrompt {
	return qualifiedPrompt{prompt: m.Prompt, context: m.Context}
}

func interval(repetitions int) int {
	return int(math.Round(math.Pow(spacedRepetitionFactor, float64(repetitions))))
}
