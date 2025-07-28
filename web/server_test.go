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

func TestHandleCreateSession(t *testing.T) {
	numProficiencyLevels := 3

	metadata := []*review.FlashcardMetadata{
		{ID: 1, Prompt: "P1", Answer: "A1"},
		{ID: 2, Prompt: "P2", Answer: "A2"},
	}

	expectedInitialSession := review.SessionMetadata{
		IsNewRound:        true,
		ProficiencyCounts: make([]int, numProficiencyLevels),
		UnreviewedCount:   len(metadata),
	}

	expectedFlashcard := review.Flashcard{Metadata: *metadata[0]}

	expectedFlashcards := []*review.Flashcard{
		{Metadata: *metadata[0]},
		{Metadata: *metadata[1]},
	}

	source := review.NewMemorySource(metadata)
	store := review.NewMemoryStore()

	server, err := New(source, store, numProficiencyLevels)
	require.NoError(t, err)

	router := server.getRouter()

	session := testCreateSession(t, router, expectedInitialSession)
	testInvalidFlashcardID(t, router, session.ID)
	testGetSession(t, router, session)
	testGetFlashcards(t, router, session.ID, expectedFlashcards)
	testNextFlashcard(t, router, session.ID, expectedFlashcard)
	testSyncFlashcards(t, router, session)

	submission := review.Submission{Answer: "A1", IsFirstGuess: true}
	expectedSubmissionResponse := SubmissionResponse{
		Session: &review.SessionMetadata{
			ID:                session.ID,
			IsNewRound:        false,
			ProficiencyCounts: []int{0, 1, 0},
			UnreviewedCount:   1,
		},
		IsCorrect: true,
	}

	testSubmitFlashcard(t, router, session.ID, 1, submission, expectedSubmissionResponse)
}

func testCreateSession(t *testing.T, router *mux.Router, expectedSession review.SessionMetadata) review.SessionMetadata {
	req := httptest.NewRequest("CREATE", "/sessions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var session review.SessionMetadata
	err := json.NewDecoder(rec.Body).Decode(&session)
	require.NoError(t, err)
	require.NotEmpty(t, session.ID)
	expectedSession.ID = session.ID
	require.Equal(t, expectedSession, session)

	return session
}

func testInvalidFlashcardID(t *testing.T, router *mux.Router, sessionID string) {
	endpoint := fmt.Sprintf("/sessions/%s/flashcards/invalid/submit", sessionID)
	req := httptest.NewRequest("POST", endpoint, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func testGetSession(t *testing.T, router *mux.Router, expectedSession review.SessionMetadata) {
	req := httptest.NewRequest("GET", "/sessions/"+expectedSession.ID, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var session review.SessionMetadata
	err := json.NewDecoder(rec.Body).Decode(&session)
	require.NoError(t, err)
	require.Equal(t, expectedSession, session)
}

func testGetFlashcards(t *testing.T, router *mux.Router, sessionID string, expectedFlashcards []*review.Flashcard) {
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

func testNextFlashcard(t *testing.T, router *mux.Router, sessionID string, expectedFlashcard review.Flashcard) {
	endpoint := fmt.Sprintf("/sessions/%s/flashcards/next", sessionID)
	req := httptest.NewRequest("POST", endpoint, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var flashcard review.Flashcard
	err := json.NewDecoder(rec.Body).Decode(&flashcard)
	require.NoError(t, err)
	require.Equal(t, expectedFlashcard, flashcard)
}

func testSyncFlashcards(t *testing.T, router *mux.Router, expectedSession review.SessionMetadata) {
	endpoint := fmt.Sprintf("/sessions/%s/flashcards/sync", expectedSession.ID)
	req := httptest.NewRequest("POST", endpoint, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var session review.SessionMetadata
	err := json.NewDecoder(rec.Body).Decode(&session)
	require.NoError(t, err)
	require.Equal(t, expectedSession, session)
}

func testSubmitFlashcard(
	t *testing.T,
	router *mux.Router,
	sessionID string,
	flashcardID int64,
	submission review.Submission,
	expectedResponse SubmissionResponse,
) {
	body, err := json.Marshal(submission)
	require.NoError(t, err)

	endpoint := fmt.Sprintf("/sessions/%s/flashcards/%d/submit", sessionID, flashcardID)
	req := httptest.NewRequest("POST", endpoint, bytes.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	var response SubmissionResponse
	err = json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	require.Equal(t, expectedResponse, response)
}
