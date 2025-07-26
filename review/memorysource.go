package review

import "context"

// MemorySource stores flashcard metadata in memory.
type MemorySource struct {
	metadata []*FlashcardMetadata
}

// ReadAll returns the metadata for all flashcards.
func (s *MemorySource) ReadAll(ctx context.Context) ([]*FlashcardMetadata, error) {
	return s.metadata, nil
}
