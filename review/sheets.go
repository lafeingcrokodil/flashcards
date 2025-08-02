package review

import (
	"context"
	"strconv"

	"github.com/lafeingcrokodil/flashcards/sheets"
)

// SheetSource stores flashcard metadata in a Google Sheets spreadsheet.
type SheetSource struct {
	// SpreadsheetID uniquely identifies the spreadsheet.
	SpreadsheetID string `json:"spreadsheetId"`
	// CellRange is the range of cells containing the data.
	CellRange string `json:"cellRange"`
	// IDHeader is the name of the column containing unique IDs.
	IDHeader string `json:"idHeader"`
	// PromptHeader is the name of the column containing the prompts.
	PromptHeader string `json:"promptHeader"`
	// ContextHeader is the name of the column containing the context (if any).
	ContextHeader string `json:"contextHeader"`
	// AnswerHeader is the name of the column containing the answers.
	AnswerHeader string `json:"answerHeader"`
}

// GetAll returns the metadata for all flashcards.
func (s *SheetSource) GetAll(ctx context.Context) ([]*FlashcardMetadata, error) {
	records, err := sheets.ReadSheet(ctx, s.SpreadsheetID, s.CellRange)
	if err != nil {
		return nil, err
	}

	metadata := make([]*FlashcardMetadata, 0, len(records))

	for _, record := range records {
		id, err := strconv.ParseInt(record[s.IDHeader], 10, 64)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, &FlashcardMetadata{
			ID:      id,
			Prompt:  record[s.PromptHeader],
			Context: record[s.ContextHeader],
			Answer:  record[s.AnswerHeader],
		})
	}

	return metadata, nil
}
