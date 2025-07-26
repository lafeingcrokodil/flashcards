package review

import (
	"math"
)

// Flashcard represents the state of a flashcard.
type Flashcard struct {
	// Metadata stores immutable data like the prompt and answer.
	Metadata FlashcardMetadata `json:"metadata"`
	// Stats stores mutable data like the view count.
	Stats FlashcardStats `json:"stats"`
}

// FlashcardMetadata stores immutable flashcard data like the prompt.
type FlashcardMetadata struct {
	// ID uniquely identifies the flashcard.
	ID int64 `json:"id"`
	// Prompt is the text to be shown to the user.
	Prompt string `json:"prompt"`
	// Context helps narrow down possible answers.
	Context string `json:"context"`
	// Answer is the accepted answer.
	Answer string `json:"answer"`
}

// FlashcardStats stores mutable flashcard data like the view count.
type FlashcardStats struct {
	// ViewCount is the number of times the flashcard has been reviewed.
	ViewCount int `json:"viewCount"`
	// Repetitions is the number of successful reviews in a row.
	Repetitions int `json:"repetitions"`
	// NextReview is the round in which the card is due to be reviewed next.
	NextReview int `json:"nextReview"`
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
