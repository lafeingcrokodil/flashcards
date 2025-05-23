package app

import (
	"encoding/json"
	"math"
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

// Submit checks if the answer is correct and updates the review session's state accordingly.
func (s *ReviewSession) Submit(answer string, isFirstGuess bool) (ok bool) {
	// Get the current flash card.
	f := s.Current[0]

	// Check if the submitted answer is correct.
	ok = f.Check(answer)

	// If this is the first guess (not a correction to an incorrect answer),
	// then we need to update the stats and move the card to the appropriate deck.
	if isFirstGuess {
		if ok {
			s.correctCount++
			if f.Proficiency < len(s.Decks)-1 {
				f.Proficiency++
			}
		} else {
			f.Proficiency = 0
		}
		f.ViewCount++
		s.viewCount++
		s.Decks[f.Proficiency] = append(s.Decks[f.Proficiency], f)
	}

	// Once the user provides the correct answer, we can select the next flashcard.
	if ok {
		// If the current round is still in progress, we can just continue.
		if len(s.Current) > 1 {
			s.Current = s.Current[1:]
			return
		}

		// Otherwise, we can start the next round by collecting flashcards from
		// any decks that are scheduled for review.
		s.RoundCount++
		s.Current = nil
		for i, deck := range s.Decks {
			if s.RoundCount%int(math.Pow(2, float64(i))) == 0 {
				var popped []Flashcard
				popped, s.Decks[i] = pop(deck, batchSize)
				s.Current = append(s.Current, popped...)
			}
		}

		// The next round will also always include some unreviewed flashcards, if any remain.
		var popped []Flashcard
		popped, s.Unreviewed = pop(s.Unreviewed, batchSize)
		s.Current = append(s.Current, popped...)
	}

	return
}

// pop removes the specified number of elements from the front of the queue.
func pop[T any](queue []T, numElemsToPop int) (popped []T, remaining []T) {
	if len(queue) > numElemsToPop {
		return queue[0:numElemsToPop], queue[numElemsToPop:]
	}
	return queue, nil
}
