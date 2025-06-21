package review

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCSVReader_Read(t *testing.T) {
	testCases := []struct {
		filepath           string
		expectedFlashcards []*Flashcard
		expectedErr        string
	}{
		{
			filepath: "testdata/unambiguous.tsv",
			expectedFlashcards: []*Flashcard{
				{ID: "1", Prompt: "P1", Context: "C1", Answer: "A1"},
				{ID: "2", Prompt: "P1", Context: "C2", Answer: "A2"},
				{ID: "3", Prompt: "P1", Answer: "A3"},
				{ID: "4", Prompt: "P2", Context: "C1", Answer: "A1"},
			},
		},
		{
			filepath:    "testdata/invalid.tsv",
			expectedErr: "open testdata/invalid.tsv: no such file or directory",
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		r := CSVReader{
			Filepath:      tc.filepath,
			Delimiter:     '\t',
			IDHeader:      "id",
			PromptHeader:  "prompt",
			ContextHeader: "context",
			AnswerHeader:  "answer",
		}
		actualFlashcards, actualErr := r.Read(ctx)
		if tc.expectedErr != "" {
			assert.EqualError(t, actualErr, tc.expectedErr)
		} else {
			assert.NoError(t, actualErr)
		}
		assert.Equal(t, tc.expectedFlashcards, actualFlashcards)
	}
}
