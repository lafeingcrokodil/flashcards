package review

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromCSV(t *testing.T) {
	testCases := []struct {
		filepath           string
		expectedFlashcards []*Flashcard
		expectedErr        string
	}{
		{
			filepath:    "testdata/ambiguous.tsv",
			expectedErr: "ambiguous answer for P1 (C1)",
		},
		{
			filepath: "testdata/unambiguous.tsv",
			expectedFlashcards: []*Flashcard{
				{ID: "3", Prompt: "P1", Answer: "A3"},
				{ID: "1", Prompt: "P1", Context: "C1", Answer: "A1"},
				{ID: "2", Prompt: "P1", Context: "C2", Answer: "A2"},
				{ID: "4", Prompt: "P2", Context: "C1", Answer: "A1"},
			},
		},
	}

	for _, tc := range testCases {
		lc := LoadConfig{
			Filepath:      tc.filepath,
			Delimiter:     '\t',
			IDHeader:      "id",
			PromptHeader:  "prompt",
			ContextHeader: "context",
			AnswerHeader:  "answer",
		}
		actualFlashcards, actualErr := LoadFromCSV(lc)
		if tc.expectedErr != "" {
			assert.EqualError(t, actualErr, tc.expectedErr)
		} else {
			assert.NoError(t, actualErr)
		}
		assert.Equal(t, tc.expectedFlashcards, actualFlashcards)
	}
}
