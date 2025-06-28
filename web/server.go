package web

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lafeingcrokodil/flashcards/review"
)

// Server is a web server for reviewing flashcards.
type Server struct {
	// session is the current review session.
	session *review.Session
	// store is where the session state will be backed up.
	store review.SessionStore
}

// Submission is data submitted by the UI.
type Submission struct {
	// Answer is the answer provided by the user.
	Answer string `json:"answer"`
	// IsFirstGuess is true if the answer is the user's first guess.
	IsFirstGuess bool `json:"isFirstGuess"`
}

// New initializes a new server.
func New(ctx context.Context, fr review.FlashcardReader, store review.SessionStore) (*Server, error) {
	s, err := review.NewSession(ctx, fr, store)
	if err != nil {
		return nil, err
	}
	return &Server{session: s, store: store}, nil
}

// Start starts the server.
func (s *Server) Start(port int) error {
	r := mux.NewRouter()
	r.HandleFunc("/state", s.getState).Methods("GET")
	r.HandleFunc("/state", s.patchState).Methods("PATCH")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public")))

	addr := fmt.Sprintf("localhost:%d", port)
	fmt.Printf("INFO\tStarting server at http://%s...\n", addr)
	return http.ListenAndServe(addr, r)
}

func (s *Server) getState(w http.ResponseWriter, _ *http.Request) {
	err := json.NewEncoder(w).Encode(s.session)
	if err != nil {
		fmt.Printf("ERROR\t%v\n", err)
	}
}

func (s *Server) patchState(w http.ResponseWriter, req *http.Request) {
	var submission Submission

	body, err := io.ReadAll(req.Body)
	if err != nil {
		fmt.Printf("ERROR\t%v\n", err)
		return
	}

	err = json.Unmarshal(body, &submission)
	if err != nil {
		fmt.Printf("ERROR\t%v\n", err)
		return
	}

	ok := s.session.Submit(submission.Answer, submission.IsFirstGuess)
	if !ok {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	err = s.store.Write(req.Context(), s.session)
	if err != nil {
		fmt.Printf("ERROR\t%v\n", err)
	}

	err = json.NewEncoder(w).Encode(s.session)
	if err != nil {
		fmt.Printf("ERROR\t%v\n", err)
	}
}
