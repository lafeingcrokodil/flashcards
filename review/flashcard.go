package review

import (
	"math"
)

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
	ID int64 `firestore:"id"`
	// Prompt is the text to be shown to the user.
	Prompt string `firestore:"prompt"`
	// Context helps narrow down possible answers.
	Context string `firestore:"context,omitempty"`
	// Answer is the accepted answer.
	Answer string `firestore:"answer"`
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

// Update updates the flashcard's stats after being reviewed.
func (f *Flashcard) Update(correct bool, round int) {
	f.Stats.ViewCount++

	if correct {
		f.Stats.NextReview = round + int(math.Round(math.Pow(2, float64(f.Stats.Repetitions))))
		f.Stats.Repetitions++
	} else {
		f.Stats.NextReview = round + 1
		f.Stats.Repetitions = 0
	}
}
