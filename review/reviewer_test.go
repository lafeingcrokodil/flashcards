package review

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReviewer_NextFlashcard(t *testing.T) {
	numFlashcards := 3
	numProficiencyLevels := 3

	expectedInitialSession := &SessionMetadata{
		IsNewRound:        true,
		ProficiencyCounts: make([]int, numProficiencyLevels),
		UnreviewedCount:   numFlashcards,
	}

	testCases := []struct {
		correct           bool
		isFirstGuess      bool
		expectedFlashcard *Flashcard
		expectedSession   *SessionMetadata
	}{
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
			},
			expectedSession: &SessionMetadata{Round: 0, IsNewRound: false, ProficiencyCounts: []int{0, 1, 0}, UnreviewedCount: 2},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
			},
			expectedSession: &SessionMetadata{Round: 1, IsNewRound: false, ProficiencyCounts: []int{0, 2, 0}, UnreviewedCount: 1},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 1},
			},
			expectedSession: &SessionMetadata{Round: 1, IsNewRound: false, ProficiencyCounts: []int{0, 1, 1}, UnreviewedCount: 1},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
			},
			expectedSession: &SessionMetadata{Round: 2, IsNewRound: false, ProficiencyCounts: []int{0, 2, 1}, UnreviewedCount: 0},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 2},
			},
			expectedSession: &SessionMetadata{Round: 2, IsNewRound: false, ProficiencyCounts: []int{0, 1, 2}, UnreviewedCount: 0},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 3},
			},
			expectedSession: &SessionMetadata{Round: 3, IsNewRound: false, ProficiencyCounts: []int{0, 1, 2}, UnreviewedCount: 0},
		},
		{
			correct:      false,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 3},
			},
			expectedSession: &SessionMetadata{Round: 3, IsNewRound: false, ProficiencyCounts: []int{0, 1, 2}, UnreviewedCount: 0},
		},
		{
			correct:      true,
			isFirstGuess: false,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(3),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 3},
			},
			expectedSession: &SessionMetadata{Round: 3, IsNewRound: false, ProficiencyCounts: []int{1, 0, 2}, UnreviewedCount: 0},
		},
	}

	ctx := context.Background()

	r := &Reviewer{source: newMemorySource(numFlashcards), store: NewMemoryStore()}

	session, err := r.CreateSession(ctx, numProficiencyLevels)
	require.NoError(t, err)
	expectedInitialSession.ID = session.ID
	require.Equal(t, expectedInitialSession, session)

	for i, tc := range testCases {
		tc.expectedSession.ID = expectedInitialSession.ID

		f, err := r.NextFlashcard(ctx, session.ID)
		require.NoError(t, err, i)
		require.Equal(t, tc.expectedFlashcard, f, i)

		var answer string
		if tc.correct {
			answer = f.Metadata.Answer
		}

		submission := &Submission{Answer: answer, IsFirstGuess: tc.isFirstGuess}

		session, ok, err := r.Submit(ctx, session.ID, f.Metadata.ID, submission)
		require.NoError(t, err, i)
		require.Equal(t, tc.correct, ok, i)
		require.Equal(t, tc.expectedSession, session, i)
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

	session, err := r.CreateSession(ctx, 5)
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
