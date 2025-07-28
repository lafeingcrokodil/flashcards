package review

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReviewer_NextFlashcard(t *testing.T) {
	numFlashcards := 2
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
			expectedSession: &SessionMetadata{Round: 0, IsNewRound: false, ProficiencyCounts: []int{0, 1, 0}, UnreviewedCount: 1},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
			},
			expectedSession: &SessionMetadata{Round: 1, IsNewRound: false, ProficiencyCounts: []int{0, 2, 0}, UnreviewedCount: 0},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 1},
			},
			expectedSession: &SessionMetadata{Round: 1, IsNewRound: false, ProficiencyCounts: []int{0, 1, 1}, UnreviewedCount: 0},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 1, Repetitions: 1, NextReview: 2},
			},
			expectedSession: &SessionMetadata{Round: 2, IsNewRound: false, ProficiencyCounts: []int{0, 0, 2}, UnreviewedCount: 0},
		},
		{
			correct:      true,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(1),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 3},
			},
			expectedSession: &SessionMetadata{Round: 3, IsNewRound: false, ProficiencyCounts: []int{0, 0, 2}, UnreviewedCount: 0},
		},
		{
			correct:      false,
			isFirstGuess: true,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 4},
			},
			expectedSession: &SessionMetadata{Round: 4, IsNewRound: true, ProficiencyCounts: []int{0, 0, 2}, UnreviewedCount: 0},
		},
		{
			correct:      true,
			isFirstGuess: false,
			expectedFlashcard: &Flashcard{
				Metadata: flashcardMetadata(2),
				Stats:    FlashcardStats{ViewCount: 2, Repetitions: 2, NextReview: 4},
			},
			expectedSession: &SessionMetadata{Round: 4, IsNewRound: false, ProficiencyCounts: []int{1, 0, 1}, UnreviewedCount: 0},
		},
	}

	ctx := context.Background()

	r := NewReviewer(newMemorySource(numFlashcards), NewMemoryStore())

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
	const initialNumFlashcards = 6
	const numProficiencyLevels = 3

	expectedSession := &SessionMetadata{
		Round:             1,
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 1, 0},
		UnreviewedCount:   5,
	}

	metadataUpdate := []*FlashcardMetadata{
		{ID: 1, Prompt: "What is 1?", Answer: "1"},               // unchanged reviewed flashcard
		{ID: 2, Prompt: "What is X?", Answer: "2"},               // flashcard with changed prompt
		{ID: 3, Prompt: "What is 3?", Answer: "X"},               // flashcard with changed answer
		{ID: 4, Prompt: "What is 4?", Answer: "4", Context: "X"}, // flashcard with changed context
		{ID: 6, Prompt: "What is 6?", Answer: "6"},               // unchanged unreviewed flashcard
		{ID: 7, Prompt: "What is 7?", Answer: "7"},               // new flashcard
	}

	expectedFlashcards := []*Flashcard{
		{Metadata: *metadataUpdate[0], Stats: flashcardStats(1)}, // stats are preserved
		{Metadata: *metadataUpdate[1]},
		{Metadata: *metadataUpdate[2]},
		{Metadata: *metadataUpdate[3]},
		{Metadata: *metadataUpdate[4]},
		{Metadata: *metadataUpdate[5]},
	}

	ctx := context.Background()

	r := NewReviewer(newMemorySource(initialNumFlashcards), NewMemoryStore())

	session, err := r.CreateSession(ctx, numProficiencyLevels)
	require.NoError(t, err)
	expectedSession.ID = session.ID

	for i := 1; i <= initialNumFlashcards-1; i++ {
		stats := flashcardStats(i)
		err := r.store.SetFlashcardStats(ctx, session.ID, int64(i), &stats)
		require.NoError(t, err, i)
	}

	err = r.store.SetSessionMetadata(ctx, session.ID, &SessionMetadata{
		ID:                session.ID,
		Round:             1,
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 1, 4},
		UnreviewedCount:   1,
	})
	require.NoError(t, err)

	r.source = NewMemorySource(metadataUpdate)

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

	source := NewMemorySource([]*FlashcardMetadata{
		{ID: 1, Prompt: "P1", Answer: "A1", Context: "C1"},
		{ID: 2, Prompt: "P1", Answer: "A2", Context: "C1"},
	})

	r := NewReviewer(source, NewMemoryStore())

	_, err := r.getFlashcardMetadata(ctx)
	require.EqualError(t, err, expectedErr)
}
