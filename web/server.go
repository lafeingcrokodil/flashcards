package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lafeingcrokodil/flashcards/review"
)

// Server is a web server for reviewing flashcards.
type Server struct {
	// Session is the current review session.
	Session *review.Session
	// BackupPath is the file path where the state will be backed up.
	BackupPath string
}

// Start starts the server.
func (s *Server) Start(addr string) error {
	http.HandleFunc("/session", s.session)
	http.Handle("/", http.FileServer(http.Dir("./public")))
	return http.ListenAndServe(addr, nil)
}

func (s *Server) session(w http.ResponseWriter, _ *http.Request) {
	b, err := json.MarshalIndent(s.Session, "", "  ")
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}

	_, err = fmt.Fprint(w, string(b))
	if err != nil {
		fmt.Printf("ERROR: %v", err)
	}
}
