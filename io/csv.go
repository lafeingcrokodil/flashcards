package io

import (
	"encoding/csv"
	"io"
	"os"
)

// ReadCSVFile parses the contents of a CSV file. The first line of the file
// must contain the column headers. The result is an array of records, one for
// each line of the file (excluding the first line), with each record mapping
// column headers to values.
func ReadCSVFile(filepath string, delimiter rune) ([]map[string]string, error) {
	var headers []string
	var records []map[string]string

	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)
	r.Comma = delimiter

	for {
		values, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if headers == nil {
			headers = values
			continue
		}
		record := make(map[string]string, len(headers))
		for i, header := range headers {
			record[header] = values[i]
		}
		records = append(records, record)
	}

	return records, nil
}
