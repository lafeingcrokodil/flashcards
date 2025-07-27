package review

import "context"

// MemorySource stores flashcard metadata in memory.
type MemorySource struct {
	metadata []*FlashcardMetadata
}

// GetAll returns the metadata for all flashcards.
func (s *MemorySource) GetAll(ctx context.Context) ([]*FlashcardMetadata, error) {
	return s.metadata, nil
}
