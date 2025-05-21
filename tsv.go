package main

import (
	"encoding/csv"
	"io"
	"os"
)

// ReadAllTSV parses the contents of a TSV file.
func ReadAllTSV(filepath string) ([]map[string]string, error) {
	var headers []string
	var records []map[string]string

	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(f)
	r.Comma = '\t'

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
