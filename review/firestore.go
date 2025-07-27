package review

import (
	"context"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// FirestoreStore stores a review session's state in a Cloud Firestore database.
type FirestoreStore struct {
	// client handles communication with the Firestore database.
	client *firestore.Client
	// collection is the name of the Firestore collection containing the sessions.
	collection string
	// sessionID uniquely identifies the review session.
	sessionID string
}

// NewFirestoreStore returns a FirestoreStore for a new or existing session.
func NewFirestoreStore(ctx context.Context, client *firestore.Client, collection, sessionID string) (*FirestoreStore, error) {
	s := &FirestoreStore{
		client:     client,
		collection: collection,
		sessionID:  sessionID,
	}

	// If no session ID is provided, create a new session.
	if s.sessionID == "" {
		s.sessionID = uuid.NewString()
		fmt.Printf("INFO\tCreating new session with ID %s\n", s.sessionID)
		err := s.SetSessionMetadata(ctx, &SessionMetadata{New: true})
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// BulkSyncFlashcards aligns the session data with the source of truth for the flashcard metadata.
func (s *FirestoreStore) BulkSyncFlashcards(ctx context.Context, metadata []*FlashcardMetadata) (err error) {
	metadataByID := make(map[int64]*FlashcardMetadata, len(metadata))
	for _, m := range metadata {
		metadataByID[m.ID] = m
	}

	writerCtx, cancelWrites := context.WithCancel(ctx)

	writer := s.client.BulkWriter(writerCtx)
	defer func() {
		if err != nil {
			cancelWrites()
		}
		writer.End()
	}()

	var existingFlashcards []*Flashcard
	existingFlashcards, err = s.GetFlashcards(ctx)
	if err != nil {
		return err
	}

	// Update and clean up existing flashcards.
	for _, e := range existingFlashcards {
		m, ok := metadataByID[e.Metadata.ID]
		if !ok {
			fmt.Printf("INFO\tRemoving flashcard with ID %d (%s)\n", e.Metadata.ID, e.Metadata.Answer)
			_, err = writer.Delete(s.flashcardRef(e.Metadata.ID))
			if err != nil {
				return
			}
			continue
		}

		if e.Metadata != *m {
			fmt.Printf("INFO\tUpdating metadata for ID %d: %v > %v\n", m.ID, e.Metadata, m)
			e.Metadata = *m
			e.Stats = FlashcardStats{}
			_, err = writer.Set(s.flashcardRef(e.Metadata.ID), e)
			if err != nil {
				return
			}
		}

		// The flashcard should now be fully synced, so no further processing is required.
		delete(metadataByID, e.Metadata.ID)
	}

	// Add any missing flashcards.
	for _, m := range metadataByID {
		fmt.Printf("INFO\tAdding flashcard with ID %d (%s)\n", m.ID, m.Answer)
		doc := s.flashcardRef(m.ID)
		_, err = writer.Set(doc, &Flashcard{Metadata: *m})
		if err != nil {
			return
		}
	}

	return
}

// GetFlashcards returns all flashcards.
func (s *FirestoreStore) GetFlashcards(ctx context.Context) ([]*Flashcard, error) {
	iter := s.sessionRef().
		Collection("flashcards").
		OrderBy("metadata.id", firestore.Asc).
		Documents(ctx)
	return s.lookupAllFlashcards(ctx, iter)
}

// NextReviewed returns a flashcard that is due to be reviewed again.
func (s *FirestoreStore) NextReviewed(ctx context.Context, round int) (*Flashcard, error) {
	iter := s.sessionRef().
		Collection("flashcards").
		Where("stats.viewCount", ">", 0).
		Where("stats.nextReview", "<=", round).
		OrderBy("stats.nextReview", firestore.Asc).
		OrderBy("stats.viewCount", firestore.Desc).
		OrderBy(firestore.DocumentID, firestore.Asc).
		Limit(1).
		Documents(ctx)
	return s.lookupFirstFlashcard(ctx, iter)
}

// NextUnreviewed returns a flashcard that has never been reviewed before.
func (s *FirestoreStore) NextUnreviewed(ctx context.Context) (*Flashcard, error) {
	iter := s.sessionRef().
		Collection("flashcards").
		Where("stats.viewCount", "==", 0).
		OrderBy(firestore.DocumentID, firestore.Asc).
		Limit(1).
		Documents(ctx)
	return s.lookupFirstFlashcard(ctx, iter)
}

// GetSessionMetadata returns the current session metadata.
func (s *FirestoreStore) GetSessionMetadata(ctx context.Context) (*SessionMetadata, error) {
	var metadata SessionMetadata

	doc, err := s.sessionRef().
		Get(ctx)
	if err != nil {
		return nil, err
	}

	err = doc.DataTo(&metadata)
	if err != nil {
		return nil, err
	}

	return &metadata, nil
}

// UpdateFlashcardStats updates a flashcard's stats.
func (s *FirestoreStore) UpdateFlashcardStats(ctx context.Context, flashcardID int64, stats *FlashcardStats) error {
	_, err := s.flashcardRef(flashcardID).
		Set(ctx, &Flashcard{Stats: *stats}, firestore.Merge([]string{"stats"}))
	return err
}

// SetSessionMetadata updates the session metadata.
func (s *FirestoreStore) SetSessionMetadata(ctx context.Context, metadata *SessionMetadata) error {
	_, err := s.sessionRef().
		Set(ctx, metadata)
	return err
}

func (s *FirestoreStore) flashcardRef(flashcardID int64) *firestore.DocumentRef {
	return s.sessionRef().
		Collection("flashcards").
		Doc(strconv.FormatInt(flashcardID, 10))
}

func (s *FirestoreStore) sessionRef() *firestore.DocumentRef {
	return s.client.Collection(s.collection).Doc(s.sessionID)
}

func (s *FirestoreStore) lookupFirstFlashcard(ctx context.Context, iter *firestore.DocumentIterator) (*Flashcard, error) {
	flashcards, err := s.lookupAllFlashcards(ctx, iter)
	if err != nil {
		return nil, err
	}

	if len(flashcards) == 0 {
		return nil, nil
	}

	return flashcards[0], nil
}

func (s *FirestoreStore) lookupAllFlashcards(ctx context.Context, iter *firestore.DocumentIterator) ([]*Flashcard, error) {
	var flashcards []*Flashcard

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var f Flashcard
		if err = doc.DataTo(&f); err != nil {
			return nil, err
		}

		flashcards = append(flashcards, &f)
	}

	return flashcards, nil
}
