package io

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SheetStore_GetAll(t *testing.T) {
	expectedRecords := []map[string]string{
		{"id": "1", "prompt": "P1", "context": "C1", "answer": "A1"},
		{"id": "2", "prompt": "P1", "context": "C2", "answer": "A2"},
		{"id": "3", "prompt": "P1", "answer": "A3"},
		{"id": "4", "prompt": "P2", "context": "C1", "answer": "A1"},
	}

	ctx := context.Background()

	spreadsheetID := os.Getenv("FLASHCARDS_SHEETS_ID")
	cellRange := os.Getenv("FLASHCARDS_SHEETS_CELL_RANGE")

	records, err := ReadSheet(ctx, spreadsheetID, cellRange)
	require.NoError(t, err)
	require.Equal(t, expectedRecords, records)
}
