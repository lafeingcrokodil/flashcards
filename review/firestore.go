package review

import (
	"context"
	"fmt"
	"strconv"

	"cloud.google.com/go/firestore"
)

// FireStore stores flashcards in a Cloud Firestore database.
type FireStore struct {
	// client handles communication with the Firestore database.
	client *firestore.Client
	// collection is the name of the Firestore collection containing the flashcards.
	collection string
}

func (s *FireStore) NextReviewed(ctx context.Context, round int) (*Flashcard, error) {
	var f Flashcard

	iter := s.client.Collection(s.collection).
		Where("stats.viewCount", ">", 0).
		Where("stats.nextReview", "==", round).
		OrderBy("stats.viewCount", firestore.Desc).
		OrderBy(firestore.DocumentID, firestore.Asc).
		Limit(1).
		Documents(ctx)

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

func (s *FireStore) NextUnreviewed(ctx context.Context) (*Flashcard, error) {
	var f Flashcard

	iter := s.client.Collection(s.collection).
		Where("stats.viewCount", "==", 0).
		OrderBy(firestore.DocumentID, firestore.Asc).
		Limit(1).
		Documents(ctx)

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

func (s *FireStore) Upsert(ctx context.Context, f *Flashcard) error {
	_, err := s.client.Collection(s.collection).
		Doc(strconv.FormatInt(f.Metadata.ID, 10)).
		Set(ctx, f)
	return err
}
