package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lafeingcrokodil/flashcards/review"
	"github.com/stretchr/testify/require"
)

// TODO: Add remaining tests.

func TestHandleCreateSession(t *testing.T) {
	expectedSession := review.SessionMetadata{
		IsNewRound:        true,
		ProficiencyCounts: []int{0, 0, 0, 0, 0},
		UnreviewedCount:   2,
	}

	source := review.NewMemorySource([]*review.FlashcardMetadata{
		{ID: 1, Prompt: "P1", Answer: "A1"},
		{ID: 2, Prompt: "P2", Answer: "A2"},
	})

	store := review.NewMemoryStore()

	server, err := New(source, store)
	require.NoError(t, err)

	router := server.getRouter()
	req := httptest.NewRequest("CREATE", "/sessions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var session review.SessionMetadata
	err = json.NewDecoder(rec.Body).Decode(&session)
	require.NoError(t, err)

	require.NotEmpty(t, session.ID)
	expectedSession.ID = session.ID
	require.Equal(t, expectedSession, session)
}
