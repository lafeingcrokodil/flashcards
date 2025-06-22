package audio

import (
	"context"

	"github.com/lafeingcrokodil/flashcards/io"
)

// SheetReader reads input text from a Google Sheets spreadsheet.
type SheetReader struct {
	// SpreadsheetID uniquely identifies the spreadsheet.
	SpreadsheetID string
	// CellRange is the range of cells containing the data.
	CellRange string
	// ColumnName is the name of the column containing the text to be converted to speech.
	ColumnName string
}

// Read returns a list of input strings from a Google Sheets spreadsheet.
func (r *SheetReader) Read(ctx context.Context) ([]string, error) {
	records, err := io.ReadSheet(ctx, r.SpreadsheetID, r.CellRange)
	if err != nil {
		return nil, err
	}

	inputs := make([]string, len(records))
	for _, record := range records {
		inputs = append(inputs, record[r.ColumnName])
	}

	return inputs, nil
}
