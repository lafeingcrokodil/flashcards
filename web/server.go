package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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
	http.HandleFunc("/state", s.state)
	http.HandleFunc("/submit", s.submit)
	http.Handle("/", http.FileServer(http.Dir("./public")))
	return http.ListenAndServe(addr, nil)
}

func (s *Server) state(w http.ResponseWriter, _ *http.Request) {
	b, err := json.MarshalIndent(s.Session, "", "  ")
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		_, err := fmt.Fprint(w, "")
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
		return
	}

	_, err = fmt.Fprint(w, string(b))
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func (s *Server) submit(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	answer := query.Get("answer")
	isFirstGuess, err := stringToBool(query.Get("isFirstGuess"))
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		_, err := fmt.Fprint(w, "")
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
		return
	}

	ok := s.Session.Submit(answer, isFirstGuess)
	if ok {
		err := s.saveToFile()
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
		}
	}

	_, err = fmt.Fprintf(w, "%v", ok)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}

func (s *Server) saveToFile() error {
	b, err := json.MarshalIndent(s.Session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.BackupPath, b, 0600)
}

func stringToBool(s string) (bool, error) {
	switch s {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean string: %s", s)
	}
}
