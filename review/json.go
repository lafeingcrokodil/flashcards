package review

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	fio "github.com/lafeingcrokodil/flashcards/io"
)

// LocalJSONStore stores session data in a local JSON file.
type LocalJSONStore struct {
	// Path is the path to the JSON file.
	Path string
}

// Load loads an existing session's data from a JSON file in the local file system.
// Returns nil if the file doesn't exist.
func (store *LocalJSONStore) Load(_ context.Context) (*Session, error) {
	_, err := os.Stat(store.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	b, err := os.ReadFile(store.Path)
	if err != nil {
		return nil, err
	}

	var s Session
	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Write overwrites the stored session data with the latest state.
func (store *LocalJSONStore) Write(_ context.Context, s *Session) error {
	return fio.WriteJSONFile(store.Path, s)
}
