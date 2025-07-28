package review

import "context"

// MemorySource stores flashcard metadata in memory.
type MemorySource struct {
	metadata []*FlashcardMetadata
}

// NewMemorySource returns a new MemorySource with the specified flashcards.
func NewMemorySource(metadata []*FlashcardMetadata) *MemorySource {
	return &MemorySource{metadata: metadata}
}

// GetAll returns the metadata for all flashcards.
func (s *MemorySource) GetAll(_ context.Context) ([]*FlashcardMetadata, error) {
	return s.metadata, nil
}
