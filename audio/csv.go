package audio

import (
	"context"

	"github.com/lafeingcrokodil/flashcards/io"
)

// CSVReader reads input text from a CSV file.
type CSVReader struct {
	// CSVPath is the path to the CSV file containing the text to be converted to speech.
	CSVPath string
	// Delimiter is the character that separates values in the file.
	Delimiter rune
	// ColumnName is the name of the column containing the text to be converted to speech.
	ColumnName string
}

// Read returns a list of input strings from a CSV file.
func (r *CSVReader) Read(_ context.Context) ([]string, error) {
	records, err := io.ReadCSVFile(r.CSVPath, r.Delimiter)
	if err != nil {
		return nil, err
	}

	inputs := make([]string, len(records))
	for _, record := range records {
		inputs = append(inputs, record[r.ColumnName])
	}

	return inputs, nil
}
