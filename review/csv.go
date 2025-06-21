package review

import (
	"context"

	"github.com/lafeingcrokodil/flashcards/io"
)

// CSVReader reads flashcard data from a CSV file.
type CSVReader struct {
	// Filepath is the path to the file.
	Filepath string
	// Delimiter is the character that separates values in the file.
	Delimiter rune
	// IDHeader is the name of the column containing unique IDs.
	IDHeader string
	// PromptHeader is the name of the column containing the prompts.
	PromptHeader string
	// ContextHeader is the name of the column containing the context (if any).
	ContextHeader string
	// AnswerHeader is the name of the column containing the answers.
	AnswerHeader string
}

// Read reads flashcards from a CSV file.
func (r *CSVReader) Read(_ context.Context) ([]*Flashcard, error) {
	var fs []*Flashcard

	records, err := io.ReadCSVFile(r.Filepath, r.Delimiter)
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
