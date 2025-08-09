package sheets

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SheetStore_GetAll(t *testing.T) {
	spreadsheetID := os.Getenv("FLASHCARDS_SHEETS_ID")
	cellRange := os.Getenv("FLASHCARDS_SHEETS_CELL_RANGE")

	testCases := []struct {
		id              string
		spreadsheetID   string
		cellRange       string
		expectedRecords []map[string]string
		expectedErr     string
	}{
		{
			id:            "Existing cell range",
			spreadsheetID: spreadsheetID,
			cellRange:     cellRange,
			expectedRecords: []map[string]string{
				{"id": "1", "prompt": "P1", "context": "C1", "answer": "A1"},
				{"id": "2", "prompt": "P1", "context": "C2", "answer": "A2"},
				{"id": "3", "prompt": "P1", "answer": "A3"},
				{"id": "4", "prompt": "P2", "context": "C1", "answer": "A1"},
			},
		},
		{
			id:            "Nonexistent cell range",
			spreadsheetID: spreadsheetID,
			cellRange:     "Nonexistent!A:D",
			expectedErr:   "googleapi: Error 400: Unable to parse range: Nonexistent!A:D, badRequest",
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			records, err := ReadSheet(ctx, tc.spreadsheetID, tc.cellRange)
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedRecords, records)
			}
		})
	}
}
