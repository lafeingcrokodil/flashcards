package review

// MemoryStore stores flashcards in memory. It's unoptimized and mainly
// intended for use in tests.
type MemoryStore struct {
	flashcards []*Flashcard
}

func (s *MemoryStore) NextReviewed(round int) (*Flashcard, error) {
	for _, f := range s.flashcards {
		if f.Stats.ViewCount > 0 && f.Stats.NextReview == round {
			return f, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) NextUnreviewed() (*Flashcard, error) {
	for _, f := range s.flashcards {
		if f.Stats.ViewCount == 0 {
			return f, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) Update(newState *Flashcard) error {
	for i, f := range s.flashcards {
		if f.Metadata.ID == newState.Metadata.ID {
			s.flashcards[i] = newState
		}
	}
	return nil
}
