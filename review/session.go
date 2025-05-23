package review

import (
	"encoding/json"
	"math"
	rand "math/rand/v2"
	"os"
)

const batchSize = 10
const numProficiencyLevels = 5

// Session is a flashcard review session.
type Session struct {
	// Current is the list of flashcards currently being reviewed.
	Current []*Flashcard `json:"current"`
	// Unreviewed is a list of flashcards that haven't yet been reviewed.
	Unreviewed []*Flashcard `json:"unreviewed"`
	// Decks are flashcards that have already been reviewed, grouped by accuracy.
	Decks [][]*Flashcard `json:"decks"`
	// RoundCount is the number of completed rounds.
	RoundCount int `json:"roundCount"`
	// CountByProficiency is the number of flashcards for each proficiency level.
	CountByProficiency []int `json:"countByProficiency"`
	// ViewCount is the number of flashcards that have been reviewed in this session.
	ViewCount int `json:"-"`
	// CorrectCount is the number of correct answers so far in this session.
	CorrectCount int `json:"-"`
}

// NewSession loads flashcards and initializes a new review session.
func NewSession(lc LoadConfig) (*Session, error) {
	flashcards, err := LoadFromCSV(lc)
	if err != nil {
		return nil, err
	}

	rand.Shuffle(len(flashcards), func(i, j int) {
		flashcards[i], flashcards[j] = flashcards[j], flashcards[i]
	})

	current, unreviewed := pop(flashcards, batchSize)

	return &Session{
		Current:            current,
		Unreviewed:         unreviewed,
		Decks:              make([][]*Flashcard, numProficiencyLevels),
		CountByProficiency: make([]int, numProficiencyLevels),
	}, nil
}

// LoadSession initializes a new review session picking up from where a
// previous review session left off.
func LoadSession(backupPath string) (*Session, error) {
	b, err := os.ReadFile(backupPath)
	if err != nil {
		return nil, err
	}

	var s Session
	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Submit checks if the answer is correct and updates the review session's state accordingly.
func (s *Session) Submit(answer string, isFirstGuess bool) (ok bool) {
	// Get the current flash card.
	f := s.Current[0]

	// Check if the submitted answer is correct.
	ok = f.Check(answer)

	// If this is the first guess (not a correction to an incorrect answer),
	// then we need to update the stats and move the card to the appropriate deck.
	if isFirstGuess {
		if f.ViewCount > 0 {
			s.CountByProficiency[f.Proficiency]--
		}
		if ok {
			s.CorrectCount++
			if f.Proficiency < len(s.Decks)-1 {
				f.Proficiency++
			}
		} else {
			f.Proficiency = 0
		}
		s.CountByProficiency[f.Proficiency]++
		f.ViewCount++
		s.ViewCount++
		s.Decks[f.Proficiency] = append(s.Decks[f.Proficiency], f)
	}

	// Once the user provides the correct answer, we can select the next flashcard.
	if ok {
		// Remove the most recently reviewed flashcard from the current deck.
		_, s.Current = pop(s.Current, 1)

		// If the current round is still in progress, we can just continue.
		if len(s.Current) > 0 {
			return
		}

		// Otherwise, we need to start the next round and add more flashcards to the current deck.
		var replenishOK bool
		for !replenishOK {
			replenishOK = s.replenishCurrentDeck()
		}
	}

	return
}

// replenishCurrentDeck adds cards from other decks to the current deck.
func (s *Session) replenishCurrentDeck() bool {
	var popped []*Flashcard

	s.RoundCount++

	// Collect flashcards from any decks that are scheduled for review.
	for i, deck := range s.Decks {
		if s.RoundCount%int(math.Pow(2, float64(i))) == 0 {
			popped, s.Decks[i] = pop(deck, batchSize)
			s.Current = append(s.Current, popped...)
		}
	}

	// Also include some unreviewed flashcards, if any remain.
	popped, s.Unreviewed = pop(s.Unreviewed, batchSize)
	s.Current = append(s.Current, popped...)

	return len(s.Current) > 0
}

// pop removes the specified number of elements from the front of the queue.
func pop[T any](queue []T, numElemsToPop int) (popped []T, remaining []T) {
	if len(queue) > numElemsToPop {
		return queue[0:numElemsToPop], queue[numElemsToPop:]
	}
	return queue, nil
}
