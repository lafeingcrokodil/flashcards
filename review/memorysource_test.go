package review

func newMemorySource(numFlashcards int) *MemorySource {
	memSource := NewMemorySource(nil)

	for i := 1; i <= numFlashcards; i++ {
		metadata := flashcardMetadata(i)
		memSource.metadata = append(memSource.metadata, &metadata)
	}

	return memSource
}
