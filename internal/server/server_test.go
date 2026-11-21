package server_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lafeingcrokodil/flashcards/v3/internal/server"
)

// serve sends a request to the app's handler and returns the recorded
// response.
func serve(t *testing.T, method, path string) *httptest.ResponseRecorder {
	t.Helper()

	handler := server.New(slog.New(slog.DiscardHandler))
	req := httptest.NewRequestWithContext(t.Context(), method, path, http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	return rec
}

func TestHealthz(t *testing.T) {
	t.Parallel()

	res := serve(t, http.MethodGet, "/healthz")

	if res.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", res.Code, http.StatusOK)
	}

	wantType := "text/plain; charset=utf-8"
	if got := res.Header().Get("Content-Type"); got != wantType {
		t.Errorf("Content-Type = %q, want %q", got, wantType)
	}

	if body := res.Body.String(); body != "ok" {
		t.Errorf("body = %q, want %q", body, "ok")
	}
}

func TestStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{
			name:       "healthz with wrong method",
			method:     http.MethodPost,
			path:       "/healthz",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "unknown path",
			method:     http.MethodGet,
			path:       "/unknown",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			res := serve(t, test.method, test.path)

			if res.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", res.Code, test.wantStatus)
			}
		})
	}
}
