package review

func newMemorySource(maxID int) *MemorySource {
	memSource := NewMemorySource(nil)

	for i := 1; i <= maxID; i++ {
		metadata := flashcardMetadata(i)
		memSource.metadata = append(memSource.metadata, &metadata)
	}

	return memSource
}
