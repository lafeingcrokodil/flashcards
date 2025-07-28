package review

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReviewer_Next(t *testing.T) {
	expectedInitialSession := &SessionMetadata{
		IsNewRound:        true,
		ProficiencyCounts: make([]int, 5),
		UnreviewedCount:   3,
	}

	expectedStates := []struct {
		session   *SessionMetadata
		flashcard *Flashcard
	}{
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
			},
			session: &SessionMetadata{Round: 0, IsNewRound: false, ProficiencyCounts: []int{0, 1, 0, 0, 0}, UnreviewedCount: 2},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
			},
			session: &SessionMetadata{Round: 1, IsNewRound: false, ProficiencyCounts: []int{0, 2, 0, 0, 0}, UnreviewedCount: 1},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 1},
			},
			session: &SessionMetadata{Round: 1, IsNewRound: false, ProficiencyCounts: []int{0, 1, 1, 0, 0}, UnreviewedCount: 1},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
			},
			session: &SessionMetadata{Round: 2, IsNewRound: false, ProficiencyCounts: []int{0, 2, 1, 0, 0}, UnreviewedCount: 0},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 2},
			},
			session: &SessionMetadata{Round: 2, IsNewRound: false, ProficiencyCounts: []int{0, 1, 2, 0, 0}, UnreviewedCount: 0},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 3},
			},
			session: &SessionMetadata{Round: 3, IsNewRound: false, ProficiencyCounts: []int{0, 1, 1, 1, 0}, UnreviewedCount: 0},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 3},
			},
			session: &SessionMetadata{Round: 3, IsNewRound: false, ProficiencyCounts: []int{0, 0, 2, 1, 0}, UnreviewedCount: 0},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 4},
			},
			session: &SessionMetadata{Round: 4, IsNewRound: false, ProficiencyCounts: []int{0, 0, 1, 2, 0}, UnreviewedCount: 0},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 5},
			},
			session: &SessionMetadata{Round: 5, IsNewRound: false, ProficiencyCounts: []int{0, 0, 0, 3, 0}, UnreviewedCount: 0},
		},
		// No cards to review in round 6.
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 3, Repetitions: 3, NextReview: 7},
			},
			session: &SessionMetadata{Round: 7, IsNewRound: false, ProficiencyCounts: []int{0, 0, 0, 2, 1}, UnreviewedCount: 0},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 3, Repetitions: 3, NextReview: 8},
			},
			session: &SessionMetadata{Round: 8, IsNewRound: false, ProficiencyCounts: []int{0, 0, 0, 1, 2}, UnreviewedCount: 0},
		},
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 3, Repetitions: 3, NextReview: 9},
			},
			session: &SessionMetadata{Round: 9, IsNewRound: false, ProficiencyCounts: []int{0, 0, 0, 0, 3}, UnreviewedCount: 0},
		},
		// No cards to review in rounds 10-14.
		{
			flashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 4, Repetitions: 4, NextReview: 15},
			},
			session: &SessionMetadata{Round: 15, IsNewRound: false, ProficiencyCounts: []int{0, 0, 0, 0, 3}, UnreviewedCount: 0},
		},
	}

	ctx := context.Background()

	r := &Reviewer{source: newMemorySource(3), store: NewMemoryStore()}

	session, err := r.CreateSession(ctx)
	require.NoError(t, err)
	expectedInitialSession.ID = session.ID
	require.Equal(t, expectedInitialSession, session)

	for i, expected := range expectedStates {
		expected.session.ID = expectedInitialSession.ID

		f, err := r.NextFlashcard(ctx, session.ID)
		require.NoError(t, err, i)
		require.Equal(t, expected.flashcard, f, i)

		session, ok, err := r.Submit(ctx, session.ID, f.Metadata.ID, &Submission{Answer: f.Metadata.Answer, IsFirstGuess: true})
		require.NoError(t, err, i)
		require.True(t, ok, i)
		require.Equal(t, expected.session, session, i)
	}
}

func TestReviewer_SyncFlashcards(t *testing.T) {
	expectedSession := &SessionMetadata{
		Round:             1,
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 1, 0, 0, 0},
		UnreviewedCount:   4,
	}

	expectedFlashcards := []*Flashcard{
		{
			Metadata: FlashcardMetadata{ID: 1, Prompt: "What is 1?", Answer: "1"},
			Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 1},
		},
		{
			Metadata: FlashcardMetadata{ID: 2, Prompt: "What is B?", Answer: "2"},
		},
		{
			Metadata: FlashcardMetadata{ID: 3, Prompt: "What is 3?", Answer: "C"},
		},
		{
			Metadata: FlashcardMetadata{ID: 4, Prompt: "What is 4?", Answer: "4", Context: "ctx"},
		},
		{
			Metadata: FlashcardMetadata{ID: 6, Prompt: "What is 6?", Answer: "6"},
		},
	}

	const initialFlashcardCount = 5

	ctx := context.Background()

	r := &Reviewer{source: newMemorySource(initialFlashcardCount), store: NewMemoryStore()}

	session, err := r.CreateSession(ctx)
	require.NoError(t, err)
	expectedSession.ID = session.ID

	for i := 1; i <= initialFlashcardCount; i++ {
		stats := flashcardStats(i)
		err := r.store.SetFlashcardStats(ctx, session.ID, int64(i), &stats)
		require.NoError(t, err, i)
	}

	err = r.store.SetSessionMetadata(ctx, session.ID, &SessionMetadata{
		ID:                session.ID,
		Round:             1,
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 1, 1, 1, 2},
	})
	require.NoError(t, err)

	r.source = &MemorySource{
		metadata: []*FlashcardMetadata{
			{ID: 1, Prompt: "What is 1?", Answer: "1"},
			{ID: 2, Prompt: "What is B?", Answer: "2"},
			{ID: 3, Prompt: "What is 3?", Answer: "C"},
			{ID: 4, Prompt: "What is 4?", Answer: "4", Context: "ctx"},
			{ID: 6, Prompt: "What is 6?", Answer: "6"},
		},
	}

	updatedSession, err := r.SyncFlashcards(ctx, session.ID)
	require.NoError(t, err)
	require.Equal(t, expectedSession, updatedSession)

	flashcards, err := r.GetFlashcards(ctx, session.ID)
	require.NoError(t, err)
	require.Equal(t, expectedFlashcards, flashcards)

	unchangedSession, err := r.GetSession(ctx, session.ID)
	require.NoError(t, err)
	require.Equal(t, expectedSession, unchangedSession)
}

func TestNewReviewer_getFlashcardMetadata(t *testing.T) {
	expectedErr := "answers A1 and A2 for prompt P1: answers are ambiguous"

	ctx := context.Background()

	source := &MemorySource{
		metadata: []*FlashcardMetadata{
			{ID: 1, Prompt: "P1", Answer: "A1", Context: "C1"},
			{ID: 2, Prompt: "P1", Answer: "A2", Context: "C1"},
		},
	}

	r := &Reviewer{source: source, store: NewMemoryStore()}

	_, err := r.getFlashcardMetadata(ctx)
	require.EqualError(t, err, expectedErr)
}
