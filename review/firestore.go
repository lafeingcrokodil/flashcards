package review

import (
	"context"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

// FireStore stores a review session's state in a Cloud Firestore database.
type FireStore struct {
	// client handles communication with the Firestore database.
	client *firestore.Client
	// collection is the name of the Firestore collection containing the sessions.
	collection string
	// sessionID uniquely identifies the review session.
	sessionID string
}

// NewFireStore returns a FireStore for a new or existing session.
func NewFireStore(ctx context.Context, client *firestore.Client, collection, sessionID string) (*FireStore, error) {
	s := &FireStore{
		client:     client,
		collection: collection,
		sessionID:  sessionID,
	}

	// If no session ID is provided, create a new session.
	if s.sessionID == "" {
		s.sessionID = uuid.NewString()
		fmt.Printf("INFO\tCreating new session with ID %s\n", s.sessionID)
		err := s.UpdateSessionStats(ctx, &SessionStats{New: true})
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// BulkSyncFlashcards aligns the session data with the source of truth for the flashcard metadata.
func (s *FireStore) BulkSyncFlashcards(ctx context.Context, metadata []*FlashcardMetadata) (err error) {
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

	iter := s.client.Collection(s.collection).
		Doc(s.sessionID).
		Collection("flashcards").
		Documents(ctx)

	// Update and clean up existing flashcards.
	for {
		var doc *firestore.DocumentSnapshot
		doc, err = iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}

		var e Flashcard
		if err = doc.DataTo(&e); err != nil {
			return
		}

		m, ok := metadataByID[e.Metadata.ID]
		if !ok {
			fmt.Printf("INFO\tRemoving flashcard with ID %d (%s)\n", e.Metadata.ID, e.Metadata.Answer)
			_, err = writer.Delete(doc.Ref)
			if err != nil {
				return
			}
		}

		if e.Metadata != *m {
			fmt.Printf("INFO\tUpdating metadata for ID %d: %v > %v\n", m.ID, e.Metadata, m)
			e.Metadata = *m
			e.Stats = FlashcardStats{}
			_, err = writer.Set(doc.Ref, e)
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

// NextReviewed returns a flashcard that is due to be reviewed again.
func (s *FireStore) NextReviewed(ctx context.Context, round int) (*Flashcard, error) {
	q := s.client.Collection(s.collection).
		Doc(s.sessionID).
		Collection("flashcards").
		Where("stats.viewCount", ">", 0).
		Where("stats.nextReview", "==", round).
		OrderBy("stats.viewCount", firestore.Desc).
		OrderBy(firestore.DocumentID, firestore.Asc).
		Limit(1)
	return s.lookupFlashcard(ctx, q)
}

// NextUnreviewed returns a flashcard that has never been reviewed before.
func (s *FireStore) NextUnreviewed(ctx context.Context) (*Flashcard, error) {
	q := s.client.Collection(s.collection).
		Doc(s.sessionID).
		Collection("flashcards").
		Where("stats.viewCount", "==", 0).
		OrderBy(firestore.DocumentID, firestore.Asc).
		Limit(1)
	return s.lookupFlashcard(ctx, q)
}

// SessionStats returns the current session stats.
func (s *FireStore) SessionStats(ctx context.Context) (*SessionStats, error) {
	var stats SessionStats

	doc, err := s.client.Collection(s.collection).
		Doc(s.sessionID).
		Get(ctx)
	if err != nil {
		return nil, err
	}

	err = doc.DataTo(&stats)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize session stats: %w", err)
	}

	return &stats, nil
}

// UpdateFlashcardStats updates a flashcard's stats.
func (s *FireStore) UpdateFlashcardStats(ctx context.Context, flashcardID int64, stats *FlashcardStats) error {
	_, err := s.client.Collection(s.collection).
		Doc(s.sessionID).
		Collection("flashcards").
		Doc(strconv.FormatInt(flashcardID, 10)).
		Set(ctx, &Flashcard{Stats: *stats}, firestore.Merge([]string{"stats"}))
	return err
}

// UpdateSessionStats updates the session stats.
func (s *FireStore) UpdateSessionStats(ctx context.Context, stats *SessionStats) error {
	_, err := s.client.Collection(s.collection).
		Doc(s.sessionID).
		Set(ctx, stats)
	return err
}

func (s *FireStore) flashcardRef(flashcardID int64) *firestore.DocumentRef {
	return s.client.Collection(s.collection).
		Doc(s.sessionID).
		Collection("flashcards").
		Doc(strconv.FormatInt(flashcardID, 10))
}

func (s *FireStore) lookupFlashcard(ctx context.Context, q firestore.Query) (*Flashcard, error) {
	var f Flashcard

	iter := q.Documents(ctx)

	docs, err := iter.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to lookup flashcard: %w", err)
	}

	if len(docs) == 0 {
		return nil, nil
	}

	err = docs[0].DataTo(&f)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize flashcard: %w", err)
	}

	return &f, nil
}
