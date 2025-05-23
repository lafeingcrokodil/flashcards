package app

import (
	"encoding/json"
	rand "math/rand/v2"
	"os"
)

const numProficiencyLevels = 5

// ReviewSession is a flashcard review session.
type ReviewSession struct {
	// Current is the list of flashcards currently being reviewed.
	Current []Flashcard `json:"current"`
	// Unreviewed is a list of flashcards that haven't yet been reviewed.
	Unreviewed []Flashcard `json:"unreviewed"`
	// Decks are flashcards that have already been reviewed, grouped by accuracy.
	Decks [][]Flashcard `json:"decks"`
	// RoundCount is the number of completed rounds.
	RoundCount int `json:"roundCount"`
	// ViewCount is the number of flashcards that have been reviewed in this session.
	viewCount int
	// CorrectCount is the number of correct answers so far in this session.
	correctCount int
}

// NewReviewSession loads flashcards and initializes a new review session.
func NewReviewSession(lc LoadConfig) (*ReviewSession, error) {
	flashcards, err := LoadFromCSV(lc)
	if err != nil {
		return nil, err
	}

	rand.Shuffle(len(flashcards), func(i, j int) {
		flashcards[i], flashcards[j] = flashcards[j], flashcards[i]
	})

	current, unreviewed := pop(flashcards, batchSize)

	return &ReviewSession{
		Current:    current,
		Unreviewed: unreviewed,
		Decks:      make([][]Flashcard, numProficiencyLevels),
	}, nil
}

// LoadReviewSession initializes a new review session picking up from where a
// previous review session left off.
func LoadReviewSession(backupPath string) (*ReviewSession, error) {
	b, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, err
	}

	var session ReviewSession
	err = json.Unmarshal(b, &session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// pop removes the specified number of elements from the front of the queue.
func pop[T any](queue []T, numElemsToPop int) (popped []T, remaining []T) {
	if len(queue) > numElemsToPop {
		return queue[0:numElemsToPop], queue[numElemsToPop:]
	}
	return queue, nil
}
