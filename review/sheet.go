package review

import (
	"context"
	"strconv"

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

// SheetStore stores session data in a Google Sheets spreadsheet.
type SheetStore struct {
	// SpreadsheetID uniquely identifies the spreadsheet.
	SpreadsheetID string
	// CellRange is the range of cells containing the data.
	CellRange string
}

// Load loads an existing session's data from a spreadsheet.
// Returns nil if there's no data in the spreadsheet yet.
func (store *SheetStore) Load(ctx context.Context) (*Session, error) {
	s := Session{
		Decks:              make([][]*Flashcard, numProficiencyLevels),
		CountByProficiency: make([]int, numProficiencyLevels),
	}

	records, err := io.ReadSheet(ctx, store.SpreadsheetID, store.CellRange)
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, nil
	}

	for _, record := range records {
		proficiency, err := strconv.Atoi(record["proficiency"])
		if err != nil {
			return nil, err
		}

		viewCount, err := strconv.Atoi(record["viewCount"])
		if err != nil {
			return nil, err
		}

		roundCount, err := strconv.Atoi(record["roundCount"])
		if err != nil {
			return nil, err
		}

		f := Flashcard{
			ID:          record["id"],
			Prompt:      record["prompt"],
			Context:     record["context"],
			Answer:      record["answer"],
			Proficiency: proficiency,
			ViewCount:   viewCount,
		}

		if record["current"] == "TRUE" {
			s.Current = append(s.Current, &f)
		} else if viewCount == 0 {
			s.Unreviewed = append(s.Unreviewed, &f)
		} else {
			s.Decks[proficiency] = append(s.Decks[proficiency], &f)
		}

		if viewCount > 0 {
			s.CountByProficiency[proficiency]++
		}

		s.RoundCount = roundCount
	}

	return &s, nil
}

// Write overwrites the stored session data with the latest state.
func (store *SheetStore) Write(ctx context.Context, s *Session) error {
	rows := [][]any{
		{"id", "prompt", "context", "answer", "proficiency", "viewCount", "current", "roundCount"},
	}

	for _, f := range s.Current {
		rows = append(rows, getValues(f, true, s.RoundCount))
	}

	nonCurrent := s.Unreviewed
	for _, deck := range s.Decks {
		nonCurrent = append(nonCurrent, deck...)
	}

	for _, f := range nonCurrent {
		rows = append(rows, getValues(f, false, s.RoundCount))
	}

	return io.WriteSheet(ctx, store.SpreadsheetID, store.CellRange, rows)
}

func getValues(f *Flashcard, isCurrent bool, roundCount int) []any {
	return []any{
		f.ID,
		f.Prompt,
		f.Context,
		f.Answer,
		f.Proficiency,
		f.ViewCount,
		isCurrent,
		roundCount,
	}
}
