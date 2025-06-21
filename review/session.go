package review

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// FlashcardReader reads flashcard data from a data source.
type FlashcardReader interface {
	// Read returns a list of flashcards.
	Read(ctx context.Context) ([]*Flashcard, error)
}

// NewSession initializes a new review session.
func NewSession(ctx context.Context, fr FlashcardReader, backupPath string) (*Session, error) {
	newSession, err := loadNew(ctx, fr)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(backupPath)
	if errors.Is(err, os.ErrNotExist) {
		return newSession, nil
	}

	existingSession, err := loadExisting(backupPath)
	if err != nil {
		return nil, err
	}

	existingSession.update(newSession)

	return existingSession, nil
}

// loadNew loads flashcards and initializes a new review session from scratch.
func loadNew(ctx context.Context, fr FlashcardReader) (*Session, error) {
	// Get the raw flashcard data.
	flashcards, err := fr.Read(ctx)
	if err != nil {
		return nil, err
	}

	// Check for ambiguous qualified prompts (prompt + context).
	flashcardsByPrompt := make(map[string]*Flashcard, len(flashcards))
	for _, f := range flashcards {
		prompt := f.QualifiedPrompt()
		if _, ok := flashcardsByPrompt[prompt]; ok {
			return nil, fmt.Errorf("ambiguous answer for %s", prompt)
		}
		flashcardsByPrompt[prompt] = f
	}

	// Randomize the order of the flashcards.
	rand.Shuffle(len(flashcards), func(i, j int) {
		flashcards[i], flashcards[j] = flashcards[j], flashcards[i]
	})

	// Pick the first batch of flashcards to review.
	current, unreviewed := pop(flashcards, batchSize)

	return &Session{
		Current:            current,
		Unreviewed:         unreviewed,
		Decks:              make([][]*Flashcard, numProficiencyLevels),
		CountByProficiency: make([]int, numProficiencyLevels),
	}, nil
}

// loadExisting initializes a new review session picking up from where a
// previous review session left off.
func loadExisting(backupPath string) (*Session, error) {
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
func (s *Session) Submit(answer string, isFirstGuess bool) bool {
	// Get the current flash card.
	f := s.Current[0]

	// If the user answered incorrectly, we'll stick with the current flashcard
	// and leave the stats unchanged for now.
	if !f.Check(answer) {
		return false
	}

	// Remove the most recently reviewed flashcard from the current deck.
	_, s.Current = pop(s.Current, 1)

	// Update stats.
	if f.ViewCount > 0 {
		s.CountByProficiency[f.Proficiency]--
	}
	if isFirstGuess {
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

	// Move the most recently reviewed flashcard to the appropriate deck.
	s.Decks[f.Proficiency] = append(s.Decks[f.Proficiency], f)

	// If the current round is still in progress, we can just continue.
	if len(s.Current) > 0 {
		return true
	}

	// Otherwise, we need to start the next round and add more flashcards to the current deck.
	var replenishOK bool
	for !replenishOK {
		replenishOK = s.replenishCurrentDeck()
	}

	return true
}

// UnreviewedCount returns the number of flashcards that haven't yet been reviewed.
func (s *Session) UnreviewedCount() int {
	unreviewedCount := len(s.Unreviewed)
	for _, f := range s.Current {
		if f.ViewCount == 0 {
			unreviewedCount++
		}
	}
	return unreviewedCount
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

// update modifies this session to include exactly the set of flashcards that are
// included in the new session.
func (s *Session) update(newSession *Session) {
	// Index the flashcards to make lookups easier.
	existingFlashcards := s.flashcardsByID()
	newFlashcards := newSession.flashcardsByID()

	// Check for flashcards from the new session that don't exist in the existing session,
	// and update flashcards that already exist.
	var missingFlashcards []*Flashcard
	for _, f := range newFlashcards {
		e, ok := existingFlashcards[f.ID]
		if !ok {
			fmt.Printf("INFO\tAdding flashcard with ID %s (%s)\n", f.ID, f.Answer)
			missingFlashcards = append(missingFlashcards, f)
			continue
		}
		if e.Answer != f.Answer {
			fmt.Printf("INFO\tUpdating answer for ID %s: %s > %s\n", f.ID, e.Answer, f.Answer)
			e.Answer = f.Answer
		}
		if e.Prompt != f.Prompt {
			fmt.Printf("INFO\tUpdating prompt for ID %s (%s): %s > %s\n", f.ID, e.Answer, e.Prompt, f.Prompt)
			e.Prompt = f.Prompt
		}
		if e.Context != f.Context {
			fmt.Printf("INFO\tUpdating context for ID %s (%s): %s > %s\n", f.ID, e.Answer, e.Context, f.Context)
			e.Context = f.Context
		}
	}

	// The newly added flashcards should be on top of the pile of flashcards
	// to be reviewed next.
	unreviewed := missingFlashcards

	decks := make([][]*Flashcard, numProficiencyLevels)

	// Filter out any flashcards from the existing session that aren't included
	// in the new session, and populate the updated decks.
	for _, f := range existingFlashcards {
		if _, ok := newFlashcards[f.ID]; !ok {
			fmt.Printf("INFO\tRemoving flashcard with ID %s (%s)\n", f.ID, f.Answer)
			continue
		}
		if f.ViewCount > 0 {
			decks[f.Proficiency] = append(decks[f.Proficiency], f)
		} else {
			unreviewed = append(unreviewed, f)
		}
	}

	// Update the current round to include the top flashcards from the updated
	// unreviewed deck.
	current, unreviewed := pop(unreviewed, batchSize)

	// Apply all of the changes to the existing session.
	s.Current = current
	s.Unreviewed = unreviewed
	s.Decks = decks

	// Ensure that the current deck isn't empty.
	replenishOK := len(s.Current) > 0
	for !replenishOK {
		replenishOK = s.replenishCurrentDeck()
	}
}

// flashcardsByID returns a map of all flashcards in this session,
// indexed by their unique identifier.
func (s *Session) flashcardsByID() map[string]*Flashcard {
	allFlashcards := append(s.Current, s.Unreviewed...)
	for _, deck := range s.Decks {
		allFlashcards = append(allFlashcards, deck...)
	}

	flashcardsByID := make(map[string]*Flashcard, len(allFlashcards))
	for _, f := range allFlashcards {
		flashcardsByID[f.ID] = f
	}

	return flashcardsByID
}

// pop removes the specified number of elements from the front of the queue.
func pop[T any](queue []T, numElemsToPop int) (popped []T, remaining []T) {
	if len(queue) > numElemsToPop {
		return queue[0:numElemsToPop], queue[numElemsToPop:]
	}
	return queue, nil
}
