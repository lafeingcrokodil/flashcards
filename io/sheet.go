package io

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// ReadSheet reads the contents of a Google spreadsheet. The first line of the
// specified cell range must contain the column headers. The result is an array
// of records, one for each row of the cell range (excluding the first row),
// with each record mapping column headers to values.
func ReadSheet(ctx context.Context, spreadsheetID string, cellRange string) ([]map[string]string, error) {
	var headers []string
	var records []map[string]string

	client, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return nil, err
	}

	resp, err := client.Spreadsheets.Values.Get(spreadsheetID, cellRange).Do()
	if err != nil {
		return nil, err
	}

	for _, row := range resp.Values {
		if headers == nil {
			for _, header := range row {
				headers = append(headers, fmt.Sprintf("%v", header))
			}
			continue
		}
		record := make(map[string]string, len(headers))
		for i, header := range headers {
			if i >= len(row) {
				break
			}
			record[header] = fmt.Sprintf("%v", row[i])
		}
		records = append(records, record)
	}

	return records, nil
}

// WriteSheet writes values to a Google spreadsheet.
func WriteSheet(ctx context.Context, spreadsheetID string, cellRange string, values [][]any) error {
	client, err := sheets.NewService(ctx, option.WithScopes(sheets.SpreadsheetsScope))
	if err != nil {
		return err
	}

	_, err = client.Spreadsheets.Values.
		Clear(spreadsheetID, cellRange, &sheets.ClearValuesRequest{}).
		Do()
	if err != nil {
		return err
	}

	_, err = client.Spreadsheets.Values.
		Update(spreadsheetID, cellRange, &sheets.ValueRange{Values: values}).
		ValueInputOption("RAW").
		Do()

	return err
}
