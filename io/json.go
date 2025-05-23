package io

import (
	"encoding/json"
	"os"
)

// WriteJSONFile writes the specified data to a JSON file.
func WriteJSONFile(filepath string, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, b, 0600)
}
