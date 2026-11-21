// Package server defines the HTTP routes of the app.
package server

import (
	"net/http"
)

// New returns a handler that serves all of the app's routes.
func New() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)

	return mux
}

// handleHealth reports that the server is up. It doesn't check any
// dependencies, so that their outages don't make the server itself look
// unhealthy.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	// A failed write means that the client has gone away, so there's no one to
	// report the error to.
	_, _ = w.Write([]byte("ok\n"))
}
