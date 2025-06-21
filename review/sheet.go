package review

import (
	"context"

	"github.com/lafeingcrokodil/flashcards/io"
)

// SheetReader reads flashcard data from a Google Sheets spreadsheet.
type SheetReader struct {
	// SpreadsheetID uniquely identifies the spreadsheet.
	SpreadsheetID string
	// CellRange is the range of cells containing the data.
	CellRange string
	// IDHeader is the name of the column containing unique IDs.
	IDHeader string
	// PromptHeader is the name of the column containing the prompts.
	PromptHeader string
	// ContextHeader is the name of the column containing the context (if any).
	ContextHeader string
	// AnswerHeader is the name of the column containing the answers.
	AnswerHeader string
}

// Read reads flashcards from a Google Sheets spreadsheet.
func (r *SheetReader) Read(ctx context.Context) ([]*Flashcard, error) {
	var fs []*Flashcard

	records, err := io.ReadSheet(ctx, r.SpreadsheetID, r.CellRange)
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		fs = append(fs, &Flashcard{
			ID:      record[r.IDHeader],
			Prompt:  record[r.PromptHeader],
			Context: record[r.ContextHeader],
			Answer:  record[r.AnswerHeader],
		})
	}

	return fs, nil
}
