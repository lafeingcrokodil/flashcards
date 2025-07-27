package review

func newMemorySource(maxID int) *MemorySource {
	memSource := &MemorySource{}

	for i := 1; i <= maxID; i++ {
		metadata := flashcardMetadata(i)
		memSource.metadata = append(memSource.metadata, &metadata)
	}

	return memSource
}
