// Package server defines the HTTP routes of the app.
package server

import (
	"log/slog"
	"net/http"
)

// New returns a handler that serves all of the app's routes.
func New(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz(logger))

	return mux
}

// healthz reports that the server is up. It doesn't check any dependencies,
// so that their outages don't make the server itself look unhealthy.
func healthz(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

		_, err := w.Write([]byte("ok"))
		if err != nil {
			logger.ErrorContext(r.Context(), "writing healthz response", "error", err)
		}
	}
}
