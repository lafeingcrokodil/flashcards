package review

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReviewer(t *testing.T) {
	s := &MemoryStore{
		flashcards: []*Flashcard{
			{
				Metadata: FlashcardMetadata{
					ID:     1,
					Prompt: "What is A?",
					Answer: "A",
				},
			},
			{
				Metadata: FlashcardMetadata{
					ID:     2,
					Prompt: "What is B?",
					Answer: "B",
				},
			},
			{
				Metadata: FlashcardMetadata{
					ID:     3,
					Prompt: "What is C?",
					Answer: "C",
				},
			},
		},
	}

	expectedFlashcards := []*Flashcard{
		// Round 0
		{
			Metadata: FlashcardMetadata{
				ID:     1,
				Prompt: "What is A?",
				Answer: "A",
			},
		},
		// Round 1
		{
			Metadata: FlashcardMetadata{
				ID:     2,
				Prompt: "What is B?",
				Answer: "B",
			},
		},
		{
			Metadata: FlashcardMetadata{
				ID:     1,
				Prompt: "What is A?",
				Answer: "A",
			},
			Stats: FlashcardStats{
				ViewCount:   1,
				Repetitions: 1,
				NextReview:  1,
			},
		},
		// Round 2
		{
			Metadata: FlashcardMetadata{
				ID:     3,
				Prompt: "What is C?",
				Answer: "C",
			},
		},
		{
			Metadata: FlashcardMetadata{
				ID:     2,
				Prompt: "What is B?",
				Answer: "B",
			},
			Stats: FlashcardStats{
				ViewCount:   1,
				Repetitions: 1,
				NextReview:  2,
			},
		},
		// Round 3
		{
			Metadata: FlashcardMetadata{
				ID:     1,
				Prompt: "What is A?",
				Answer: "A",
			},
			Stats: FlashcardStats{
				ViewCount:   2,
				Repetitions: 2,
				NextReview:  3,
			},
		},
		{
			Metadata: FlashcardMetadata{
				ID:     3,
				Prompt: "What is C?",
				Answer: "C",
			},
			Stats: FlashcardStats{
				ViewCount:   1,
				Repetitions: 1,
				NextReview:  3,
			},
		},
		// Round 4
		{
			Metadata: FlashcardMetadata{
				ID:     2,
				Prompt: "What is B?",
				Answer: "B",
			},
			Stats: FlashcardStats{
				ViewCount:   2,
				Repetitions: 2,
				NextReview:  4,
			},
		},
		// Round 5
		{
			Metadata: FlashcardMetadata{
				ID:     3,
				Prompt: "What is C?",
				Answer: "C",
			},
			Stats: FlashcardStats{
				ViewCount:   2,
				Repetitions: 2,
				NextReview:  5,
			},
		},
		// Round 6 - no cards to review
		// Round 7
		{
			Metadata: FlashcardMetadata{
				ID:     1,
				Prompt: "What is A?",
				Answer: "A",
			},
			Stats: FlashcardStats{
				ViewCount:   3,
				Repetitions: 3,
				NextReview:  7,
			},
		},
	}

	r := NewReviewer(s)

	for i, expected := range expectedFlashcards {
		f, err := r.Next()
		assert.NoError(t, err, i)
		assert.Equal(t, expected, f, i)
		err = r.Submit(f, true)
		assert.NoError(t, err, i)
	}
}
