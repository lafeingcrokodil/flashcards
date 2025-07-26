package review

// FlashcardStore stores a set of flashcards.
type FlashcardStore interface {
	// NextReviewed returns a flashcard that is due to be reviewed again.
	NextReviewed(round int) (*Flashcard, error)
	// NextUnreviewed returns a flashcard that has never been reviewed before.
	NextUnreviewed() (*Flashcard, error)
	// Update updates the state of the specified flashcard.
	Update(f *Flashcard) error
}

// Reviewer manages the review of a set of flashcards.
type Reviewer struct {
	// store stores the flashcards to be reviewed.
	store FlashcardStore
	// round identifies the current round, starting with 0 and incrementing from there.
	round int
	// new is true if and only if the round has just started.
	new bool
}

// NewReviewer returns a new reviewer that uses the specified flashcard store.
func NewReviewer(store FlashcardStore) *Reviewer {
	return &Reviewer{
		store: store,
		new:   true,
	}
}

// Next returns the next flashcard to be reviewed.
func (r *Reviewer) Next() (*Flashcard, error) {
	if r.new {
		r.new = false
		f, err := r.store.NextUnreviewed()
		if f != nil || err != nil {
			return f, err
		}
	}

	f, err := r.store.NextReviewed(r.round)
	if f != nil || err != nil {
		return f, err
	}

	r.round++
	r.new = true

	return r.Next()
}

// Submit updates a flashcard's state based on whether it was answered correctly or not.
func (r *Reviewer) Submit(f *Flashcard, correct bool) error {
	f.Update(correct, r.round)
	return r.store.Update(f)
}
