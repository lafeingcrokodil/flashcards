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

var (
	// ErrMissingFlashcardID is thrown if a request is missing a flashcard ID.
	ErrMissingFlashcardID = errors.New("missing flashcard ID")
	// ErrMissingSessionID is thrown if a request is missing a session ID.
	ErrMissingSessionID = errors.New("missing session ID")
)

// Server is a web server for reviewing flashcards.
type Server struct {
	reviewer             *review.Reviewer
	numProficiencyLevels int
}

// New initializes a new server.
func New(store review.SessionStore, numProficiencyLevels int) (*Server, error) {
	return &Server{
		reviewer:             review.NewReviewer(store),
		numProficiencyLevels: numProficiencyLevels,
	}, nil
}

// Start starts the server.
func (s *Server) Start(port int) error {
	router := s.getRouter()
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf("INFO\tStarting server at http://%s...\n", addr)
	return http.ListenAndServe(addr, router)
}

func (s *Server) getRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/sessions", s.handleCreateSession).Methods("POST")
	r.HandleFunc("/sessions", s.handleGetSessions).Methods("GET")
	r.HandleFunc("/sessions/{sid}", s.handleGetSession).Methods("GET")
	r.HandleFunc("/sessions/{sid}/flashcards", s.handleGetFlashcards).Methods("GET")
	r.HandleFunc("/sessions/{sid}/flashcards/next", s.handleNextFlashcard).Methods("POST")
	r.HandleFunc("/sessions/{sid}/flashcards/sync", s.handleSyncFlashcards).Methods("POST")
	r.HandleFunc("/sessions/{sid}/flashcards/{fid}/submit", s.handleSubmitFlashcard).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))
	return r
}

func (s *Server) handleCreateSession(w http.ResponseWriter, req *http.Request) {
	var source review.SheetSource
	err := json.NewDecoder(req.Body).Decode(&source)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	session, err := s.reviewer.CreateSession(req.Context(), &source, s.numProficiencyLevels)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}

	sendResponse(w, http.StatusCreated, session)
}

func (s *Server) handleGetSessions(w http.ResponseWriter, req *http.Request) {
	sessions, err := s.reviewer.GetSessions(req.Context())
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, http.StatusOK, sessions)
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
	sendResponse(w, http.StatusOK, session)
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
	sendResponse(w, http.StatusOK, flashcards)
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
	sendResponse(w, http.StatusOK, flashcard)
}

func (s *Server) handleSyncFlashcards(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sessionID, ok := vars["sid"]
	if !ok {
		sendError(w, http.StatusBadRequest, ErrMissingSessionID)
		return
	}

	var source review.SheetSource
	err := json.NewDecoder(req.Body).Decode(&source)
	if err != nil {
		sendError(w, http.StatusBadRequest, err)
		return
	}

	session, err := s.reviewer.SyncFlashcards(req.Context(), sessionID, &source)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err)
		return
	}
	sendResponse(w, http.StatusOK, session)
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
	if ok {
		sendResponse(w, http.StatusOK, session)
	} else {
		sendResponse(w, http.StatusNotModified, nil)
	}
}

func sendError(w http.ResponseWriter, statusCode int, err error) {
	fmt.Printf("ERROR\t%v\n", err)
	http.Error(w, err.Error(), statusCode)
}

func sendResponse(w http.ResponseWriter, statusCode int, data any) {
	if statusCode == http.StatusNotModified {
		w.WriteHeader(statusCode)
		return
	}

	// We're taking a calculated risk here of assuming that the encoding will
	// never fail, so we don't bother implementing the error handling in a way
	// that would allow sending an error response to the client. We just add
	// some simple logging.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Printf("ERROR\t%v\n", err)
	}
}
