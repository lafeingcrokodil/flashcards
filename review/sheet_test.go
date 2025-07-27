package review

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SheetStore_ReadAll(t *testing.T) {
	expectedFlashcards := []*FlashcardMetadata{
		{ID: 1, Prompt: "P1", Context: "C1", Answer: "A1"},
		{ID: 2, Prompt: "P1", Context: "C2", Answer: "A2"},
		{ID: 3, Prompt: "P1", Answer: "A3"},
		{ID: 4, Prompt: "P2", Context: "C1", Answer: "A1"},
	}

	s := SheetSource{
		SpreadsheetID: os.Getenv("FLASHCARDS_SHEETS_ID"),
		CellRange:     os.Getenv("FLASHCARDS_SHEETS_CELL_RANGE"),
		IDHeader:      os.Getenv("FLASHCARDS_SHEETS_ID_HEADER"),
		PromptHeader:  os.Getenv("FLASHCARDS_SHEETS_PROMPT_HEADER"),
		ContextHeader: os.Getenv("FLASHCARDS_SHEETS_CONTEXT_HEADER"),
		AnswerHeader:  os.Getenv("FLASHCARDS_SHEETS_ANSWER_HEADER"),
	}

	ctx := context.Background()

	flashcards, err := s.ReadAll(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedFlashcards, flashcards)
}
