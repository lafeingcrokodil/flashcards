package review

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlashcard_Submit(t *testing.T) {
	f := &Flashcard{
		Metadata: flashcardMetadata(1),
	}

	updates := []struct {
		id            string
		submission    *Submission
		round         int
		expectedOK    bool
		expectedState *Flashcard
	}{
		{
			id:         "Correct first review",
			submission: &Submission{Answer: "1", IsFirstGuess: true},
			round:      0,
			expectedOK: true,
			expectedState: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats: FlashcardStats{
					ViewCount:   1,
					Repetitions: 1,
					NextReview:  1,
				},
			},
		},
		{
			id:         "Correct second review",
			submission: &Submission{Answer: "1", IsFirstGuess: true},
			round:      1,
			expectedOK: true,
			expectedState: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats: FlashcardStats{
					ViewCount:   2,
					Repetitions: 2,
					NextReview:  3,
				},
			},
		},
		{
			id:         "Correct third review",
			submission: &Submission{Answer: "1", IsFirstGuess: true},
			round:      3,
			expectedOK: true,
			expectedState: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats: FlashcardStats{
					ViewCount:   3,
					Repetitions: 3,
					NextReview:  7,
				},
			},
		},
		{
			id:         "Incorrect fourth review",
			submission: &Submission{Answer: "2", IsFirstGuess: true},
			round:      7,
			expectedOK: false,
			expectedState: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats: FlashcardStats{
					ViewCount:   3,
					Repetitions: 3,
					NextReview:  7,
				},
			},
		},
		{
			id:         "Correction of fourth review",
			submission: &Submission{Answer: "1", IsFirstGuess: false},
			round:      7,
			expectedOK: true,
			expectedState: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats: FlashcardStats{
					ViewCount:   4,
					Repetitions: 0,
					NextReview:  8,
				},
			},
		},
	}

	for _, update := range updates {
		ok := f.Submit(update.submission, update.round)
		require.Equal(t, update.expectedOK, ok, update.id)
		require.Equal(t, update.expectedState, f, update.id)
	}
}

func flashcardMetadata(i int) FlashcardMetadata {
	return FlashcardMetadata{
		ID:     int64(i),
		Prompt: fmt.Sprintf("What is %d?", i),
		Answer: strconv.Itoa(i),
	}
}

func flashcardStats(i int) FlashcardStats {
	return FlashcardStats{
		ViewCount:   i,
		Repetitions: i,
		NextReview:  i,
	}
}
