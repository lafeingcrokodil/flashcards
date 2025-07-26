package review

import "context"

// MemoryStore stores flashcards in memory. It's unoptimized and mainly
// intended for use in tests.
type MemoryStore struct {
	flashcards []*Flashcard
}

func (s *MemoryStore) NextReviewed(_ context.Context, round int) (*Flashcard, error) {
	for _, f := range s.flashcards {
		if f.Stats.ViewCount > 0 && f.Stats.NextReview == round {
			return f, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) NextUnreviewed(_ context.Context) (*Flashcard, error) {
	for _, f := range s.flashcards {
		if f.Stats.ViewCount == 0 {
			return f, nil
		}
	}
	return nil, nil
}

func (s *MemoryStore) Upsert(_ context.Context, f *Flashcard) error {
	var found bool
	for i, existing := range s.flashcards {
		if existing.Metadata.ID == f.Metadata.ID {
			s.flashcards[i] = f
			found = true
		}
	}
	if !found {
		s.flashcards = append(s.flashcards, f)
	}
	return nil
}
