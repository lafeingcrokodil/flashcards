// Package web implements an HTTP server for reviewing flashcards.
package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/lafeingcrokodil/flashcards/review"
)

const numProficiencyLevels = 5

var (
	// ErrMissingFlashcardID is thrown if a request is missing a flashcard ID.
	ErrMissingFlashcardID = errors.New("missing flashcard ID")
	// ErrMissingSessionID is thrown if a request is missing a session ID.
	ErrMissingSessionID = errors.New("missing session ID")
)

// SubmissionResponse is the response for a submit request.
type SubmissionResponse struct {
	// Session contains the current session metadata.
	Session *review.SessionMetadata `json:"session"`
	// IsCorrect is true if and only if the submission had a correct answer.
	IsCorrect bool `json:"isCorrect"`
}

// Server is a web server for reviewing flashcards.
type Server struct {
	reviewer *review.Reviewer
}

// New initializes a new server.
func New(source review.FlashcardMetadataSource, store review.SessionStore) (*Server, error) {
	return &Server{reviewer: review.NewReviewer(source, store)}, nil
}

// Start starts the server.
func (s *Server) Start(port int) error {
	router := s.getRouter()
	addr := fmt.Sprintf("localhost:%d", port)
	fmt.Printf("INFO\tStarting server at http://%s...\n", addr)
	return http.ListenAndServe(addr, router)
}

func (s *Server) getRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/sessions", s.handleCreateSession).Methods("CREATE")
	r.HandleFunc("/sessions/{sid}", s.handleGetSession).Methods("GET")
	r.HandleFunc("/sessions/{sid}/flashcards", s.handleGetFlashcards).Methods("GET")
	r.HandleFunc("/sessions/{sid}/flashcards/next", s.handleNextFlashcard).Methods("POST")
	r.HandleFunc("/sessions/{sid}/flashcards/sync", s.handleSyncFlashcards).Methods("POST")
	r.HandleFunc("/sessions/{sid}/flashcards/{fid}/submit", s.handleSubmitFlashcard).Methods("POST")
	return r
}

func (s *Server) handleCreateSession(w http.ResponseWriter, req *http.Request) {
	session, err := s.reviewer.CreateSession(req.Context(), numProficiencyLevels)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, session)
}

func (s *Server) handleGetSession(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sessionID, ok := vars["sid"]
	if !ok {
		sendError(w, http.StatusBadRequest, ErrMissingSessionID)
		return
	}
	session, err := s.reviewer.GetSession(req.Context(), sessionID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, session)
}

func (s *Server) handleGetFlashcards(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sessionID, ok := vars["sid"]
	if !ok {
		sendError(w, http.StatusBadRequest, ErrMissingSessionID)
		return
	}
	flashcards, err := s.reviewer.GetFlashcards(req.Context(), sessionID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, flashcards)
}

func (s *Server) handleNextFlashcard(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sessionID, ok := vars["sid"]
	if !ok {
		sendError(w, http.StatusBadRequest, ErrMissingSessionID)
		return
	}
	flashcard, err := s.reviewer.NextFlashcard(req.Context(), sessionID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, flashcard)
}

func (s *Server) handleSyncFlashcards(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sessionID, ok := vars["sid"]
	if !ok {
		sendError(w, http.StatusBadRequest, ErrMissingSessionID)
		return
	}
	session, err := s.reviewer.SyncFlashcards(req.Context(), sessionID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, session)
}

func (s *Server) handleSubmitFlashcard(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	sessionID, ok := vars["sid"]
	if !ok {
		sendError(w, http.StatusBadRequest, ErrMissingSessionID)
		return
	}

	fid, ok := vars["fid"]
	if !ok {
		sendError(w, http.StatusBadRequest, ErrMissingFlashcardID)
		return
	}

	flashcardID, err := strconv.ParseInt(fid, 10, 64)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	var submission review.Submission
	err = json.NewDecoder(req.Body).Decode(&submission)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	session, ok, err := s.reviewer.Submit(req.Context(), sessionID, flashcardID, &submission)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendResponse(w, SubmissionResponse{Session: session, IsCorrect: ok})
}

func sendError(w http.ResponseWriter, code int, err error) {
	fmt.Printf("ERROR\t%v\n", err)
	http.Error(w, err.Error(), code)
}

func sendResponse(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
	}
}
