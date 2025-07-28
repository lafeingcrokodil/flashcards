package review

import (
	"context"
	"errors"
	"strconv"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// FirestoreStore stores a review session's state in a Cloud Firestore database.
type FirestoreStore struct {
	client     *firestore.Client
	collection string
}

// NewFirestoreStore returns a new FirestoreStore that stores data in the specified collection.
func NewFirestoreStore(client *firestore.Client, collection string) *FirestoreStore {
	return &FirestoreStore{client: client, collection: collection}
}

// DeleteFlashcards deletes the specified flashcards.
func (s *FirestoreStore) DeleteFlashcards(ctx context.Context, sessionID string, ids []int64) error {
	writer := s.client.BulkWriter(ctx)
	defer writer.End()

	for _, id := range ids {
		_, err := writer.Delete(s.flashcardRef(sessionID, id))
		if err != nil {
			return err
		}
	}

	return nil
}

// GetFlashcard returns the specified flashcard.
func (s *FirestoreStore) GetFlashcard(ctx context.Context, sessionID string, flashcardID int64) (*Flashcard, error) {
	doc, err := s.flashcardRef(sessionID, flashcardID).Get(ctx)
	if err != nil {
		return nil, err
	}

	var f Flashcard
	err = doc.DataTo(&f)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

// GetFlashcards returns all flashcards.
func (s *FirestoreStore) GetFlashcards(ctx context.Context, sessionID string) ([]*Flashcard, error) {
	iter := s.sessionRef(sessionID).
		Collection("flashcards").
		OrderBy("metadata.id", firestore.Asc).
		Documents(ctx)
	return s.lookupAllFlashcards(iter)
}

// SetFlashcards upserts the specified flashcards, clearing any existing stats.
func (s *FirestoreStore) SetFlashcards(ctx context.Context, sessionID string, metadata []*FlashcardMetadata) error {
	writer := s.client.BulkWriter(ctx)
	defer writer.End()

	for _, m := range metadata {
		_, err := writer.Set(s.flashcardRef(sessionID, m.ID), &Flashcard{Metadata: *m})
		if err != nil {
			return err
		}
	}

	return nil
}

// SetFlashcardStats updates a flashcard's stats.
func (s *FirestoreStore) SetFlashcardStats(ctx context.Context, sessionID string, flashcardID int64, stats *FlashcardStats) error {
	_, err := s.flashcardRef(sessionID, flashcardID).
		Set(ctx, &Flashcard{Stats: *stats}, firestore.Merge([]string{"stats"}))
	return err
}

// NextReviewed returns a flashcard that is due to be reviewed again.
func (s *FirestoreStore) NextReviewed(ctx context.Context, sessionID string, round int) (*FlashcardMetadata, error) {
	iter := s.sessionRef(sessionID).
		Collection("flashcards").
		Where("stats.viewCount", ">", 0).
		Where("stats.nextReview", "<=", round).
		OrderBy("stats.nextReview", firestore.Asc).
		OrderBy("stats.viewCount", firestore.Desc).
		OrderBy(firestore.DocumentID, firestore.Asc).
		Limit(1).
		Documents(ctx)
	return s.lookupFirstFlashcardMetadata(iter)
}

// NextUnreviewed returns a flashcard that has never been reviewed before.
func (s *FirestoreStore) NextUnreviewed(ctx context.Context, sessionID string) (*FlashcardMetadata, error) {
	iter := s.sessionRef(sessionID).
		Collection("flashcards").
		Where("stats.viewCount", "==", 0).
		OrderBy(firestore.DocumentID, firestore.Asc).
		Limit(1).
		Documents(ctx)
	return s.lookupFirstFlashcardMetadata(iter)
}

// GetSessionMetadata returns the current session metadata.
func (s *FirestoreStore) GetSessionMetadata(ctx context.Context, sessionID string) (*SessionMetadata, error) {
	var session SessionMetadata

	doc, err := s.sessionRef(sessionID).
		Get(ctx)
	if err != nil {
		return nil, err
	}

	err = doc.DataTo(&session)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// SetSessionMetadata updates the session metadata.
func (s *FirestoreStore) SetSessionMetadata(ctx context.Context, sessionID string, session *SessionMetadata) error {
	_, err := s.sessionRef(sessionID).
		Set(ctx, session)
	return err
}

func (s *FirestoreStore) flashcardRef(sessionID string, flashcardID int64) *firestore.DocumentRef {
	return s.sessionRef(sessionID).
		Collection("flashcards").
		Doc(strconv.FormatInt(flashcardID, 10))
}

func (s *FirestoreStore) sessionRef(sessionID string) *firestore.DocumentRef {
	return s.client.Collection(s.collection).Doc(sessionID)
}

func (s *FirestoreStore) lookupFirstFlashcardMetadata(iter *firestore.DocumentIterator) (*FlashcardMetadata, error) {
	flashcards, err := s.lookupAllFlashcards(iter)
	if err != nil {
		return nil, err
	}

	if len(flashcards) == 0 {
		return nil, ErrNotFound
	}

	return &flashcards[0].Metadata, nil
}

func (s *FirestoreStore) lookupAllFlashcards(iter *firestore.DocumentIterator) ([]*Flashcard, error) {
	var flashcards []*Flashcard

	for {
		doc, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}

		var f Flashcard
		err = doc.DataTo(&f)
		if err != nil {
			return nil, err
		}

		flashcards = append(flashcards, &f)
	}

	return flashcards, nil
}
