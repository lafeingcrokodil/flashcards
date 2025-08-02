package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/lafeingcrokodil/flashcards/review"
	"github.com/stretchr/testify/require"
)

func TestServer(t *testing.T) {
	numProficiencyLevels := 3

	metadata := []*review.FlashcardMetadata{
		{ID: 1, Prompt: "P1", Answer: "A1"},
		{ID: 2, Prompt: "P2", Answer: "A2"},
	}

	source := review.NewMemorySource(metadata)
	store := review.NewMemoryStore()

	server, err := New(source, store, numProficiencyLevels)
	require.NoError(t, err)

	router := server.getRouter()

	session := testCreateSession(t, router)
	testInvalidFlashcardID(t, router, session.ID)
	testGetSession(t, router, session.ID)
	testGetSessions(t, router, session.ID)
	testGetFlashcards(t, router, session.ID)
	testNextFlashcard(t, router, session.ID)
	testSyncFlashcards(t, router, session.ID)
	testSubmitFlashcard(t, router, session.ID)
}

func testCreateSession(t *testing.T, router *mux.Router) review.Session {
	expectedSession := &review.Session{
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 0, 0},
		UnreviewedCount:   2,
	}

	req := httptest.NewRequest("POST", "/sessions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var session review.Session
	err := json.NewDecoder(rec.Body).Decode(&session)
	require.NoError(t, err)
	require.NotEmpty(t, session.ID)
	expectedSession.ID = session.ID
	require.Equal(t, expectedSession, &session)

	return session
}

func testInvalidFlashcardID(t *testing.T, router *mux.Router, sessionID string) {
	endpoint := fmt.Sprintf("/sessions/%s/flashcards/invalid/submit", sessionID)
	req := httptest.NewRequest("POST", endpoint, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func testGetSession(t *testing.T, router *mux.Router, sessionID string) {
	expectedSession := &review.Session{
		ID:                sessionID,
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 0, 0},
		UnreviewedCount:   2,
	}

	req := httptest.NewRequest("GET", "/sessions/"+sessionID, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var session review.Session
	err := json.NewDecoder(rec.Body).Decode(&session)
	require.NoError(t, err)
	require.Equal(t, expectedSession, &session)
}

func testGetSessions(t *testing.T, router *mux.Router, sessionID string) {
	expectedSession := &review.Session{
		ID:                sessionID,
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 0, 0},
		UnreviewedCount:   2,
	}

	req := httptest.NewRequest("GET", "/sessions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var sessions []*review.Session
	err := json.NewDecoder(rec.Body).Decode(&sessions)
	require.NoError(t, err)
	require.Equal(t, []*review.Session{expectedSession}, sessions)
}

func testGetFlashcards(t *testing.T, router *mux.Router, sessionID string) {
	expectedFlashcards := []*review.Flashcard{
		{Metadata: review.FlashcardMetadata{ID: 1, Prompt: "P1", Answer: "A1"}},
		{Metadata: review.FlashcardMetadata{ID: 2, Prompt: "P2", Answer: "A2"}},
	}

	endpoint := fmt.Sprintf("/sessions/%s/flashcards", sessionID)
	req := httptest.NewRequest("GET", endpoint, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var flashcards []*review.Flashcard
	err := json.NewDecoder(rec.Body).Decode(&flashcards)
	require.NoError(t, err)
	require.Equal(t, expectedFlashcards, flashcards)
}

func testNextFlashcard(t *testing.T, router *mux.Router, sessionID string) {
	expectedFlashcard := &review.Flashcard{
		Metadata: review.FlashcardMetadata{ID: 1, Prompt: "P1", Answer: "A1"},
	}

	endpoint := fmt.Sprintf("/sessions/%s/flashcards/next", sessionID)
	req := httptest.NewRequest("POST", endpoint, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var flashcard review.Flashcard
	err := json.NewDecoder(rec.Body).Decode(&flashcard)
	require.NoError(t, err)
	require.Equal(t, expectedFlashcard, &flashcard)
}

func testSyncFlashcards(t *testing.T, router *mux.Router, sessionID string) {
	expectedSession := &review.Session{
		ID:                sessionID,
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 0, 0},
		UnreviewedCount:   2,
	}

	endpoint := fmt.Sprintf("/sessions/%s/flashcards/sync", sessionID)
	req := httptest.NewRequest("POST", endpoint, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var session review.Session
	err := json.NewDecoder(rec.Body).Decode(&session)
	require.NoError(t, err)
	require.Equal(t, expectedSession, &session)
}

func testSubmitFlashcard(t *testing.T, router *mux.Router, sessionID string) {
	testCases := []struct {
		id                 string
		flashcardID        int64
		submission         *review.Submission
		expectedStatusCode int
		expectedSession    *review.Session
	}{
		{
			id:                 "Incorrect answer",
			flashcardID:        1,
			submission:         &review.Submission{Answer: "X", IsFirstGuess: true},
			expectedStatusCode: 304,
		},
		{
			id:                 "Correct answer",
			flashcardID:        1,
			submission:         &review.Submission{Answer: "A1", IsFirstGuess: false},
			expectedStatusCode: 200,
			expectedSession: &review.Session{
				ID:                sessionID,
				IsNewRound:        false,
				ProficiencyCounts: []int{1, 0, 0},
				UnreviewedCount:   1,
			},
		},
	}

	for _, tc := range testCases {
		body, err := json.Marshal(tc.submission)
		require.NoError(t, err, tc.id)

		endpoint := fmt.Sprintf("/sessions/%s/flashcards/%d/submit", sessionID, tc.flashcardID)
		req := httptest.NewRequest("POST", endpoint, bytes.NewReader(body))
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		require.Equal(t, tc.expectedStatusCode, rec.Code, tc.id)

		if tc.expectedSession != nil {
			var session review.Session
			err = json.NewDecoder(rec.Body).Decode(&session)
			require.NoError(t, err, tc.id)
			require.Equal(t, tc.expectedSession, &session, tc.id)
		}
	}
}
