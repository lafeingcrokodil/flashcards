package review

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlashcard_Update(t *testing.T) {
	f := &Flashcard{
		Metadata: FlashcardMetadata{
			ID:     1,
			Prompt: "What is A?",
			Answer: "B",
		},
	}

	updates := []struct {
		id            string
		correct       bool
		round         int
		expectedState *Flashcard
	}{
		{
			id:      "Correct first review",
			correct: true,
			round:   0,
			expectedState: &Flashcard{
				Metadata: FlashcardMetadata{
					ID:     1,
					Prompt: "What is A?",
					Answer: "B",
				},
				Stats: FlashcardStats{
					ViewCount:   1,
					Repetitions: 1,
					NextReview:  1,
				},
			},
		},
		{
			id:      "Correct second review",
			correct: true,
			round:   1,
			expectedState: &Flashcard{
				Metadata: FlashcardMetadata{
					ID:     1,
					Prompt: "What is A?",
					Answer: "B",
				},
				Stats: FlashcardStats{
					ViewCount:   2,
					Repetitions: 2,
					NextReview:  3,
				},
			},
		},
		{
			id:      "Correct third review",
			correct: true,
			round:   3,
			expectedState: &Flashcard{
				Metadata: FlashcardMetadata{
					ID:     1,
					Prompt: "What is A?",
					Answer: "B",
				},
				Stats: FlashcardStats{
					ViewCount:   3,
					Repetitions: 3,
					NextReview:  7,
				},
			},
		},
		{
			id:      "Incorrect fourth review",
			correct: false,
			round:   7,
			expectedState: &Flashcard{
				Metadata: FlashcardMetadata{
					ID:     1,
					Prompt: "What is A?",
					Answer: "B",
				},
				Stats: FlashcardStats{
					ViewCount:   4,
					Repetitions: 0,
					NextReview:  8,
				},
			},
		},
	}

	for _, update := range updates {
		f.Update(update.correct, update.round)
		assert.Equal(t, update.expectedState, f, update.id)
	}
}
