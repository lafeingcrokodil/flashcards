package server_test

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/lafeingcrokodil/flashcards/v3/internal/server"
)

const (
	indexHTML = "<h1>Flashcards</h1>"
	appJS     = "console.log('app');"
	robotsTXT = "User-agent: *\nDisallow: /\n"
	// immutable is the Cache-Control header for files that never change.
	immutable = "public, max-age=31536000, immutable"
)

// newFrontend returns a small stand-in for the built front end.
func newFrontend() fstest.MapFS {
	return fstest.MapFS{
		"index.html":                       {Data: []byte(indexHTML)},
		"robots.txt":                       {Data: []byte(robotsTXT)},
		"_app/version.json":                {Data: []byte(`{"version":"1"}`)},
		"_app/immutable/entry/app.123.js":  {Data: []byte(appJS)},
		"_app/immutable/assets/dir.css":    {Mode: fs.ModeDir},
		"_app/immutable/assets/dir.css/ok": {Data: []byte("not served")},
	}
}

func TestRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		// wantBody is the expected body, or empty if the body isn't checked
		// (e.g. for errors, whose bodies come from net/http).
		wantBody string
		// wantType is the expected Content-Type, or empty if it isn't checked.
		wantType string
		// wantCache is the expected Cache-Control, where empty means none.
		wantCache string
	}{
		{
			name:       "health",
			path:       "/health",
			wantStatus: http.StatusOK,
			wantBody:   "ok\n",
			wantType:   "text/plain; charset=utf-8",
		},
		{
			name:       "health with HEAD",
			method:     http.MethodHead,
			path:       "/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "health with wrong method",
			method:     http.MethodPost,
			path:       "/health",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "root",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
			wantType:   "text/html; charset=utf-8",
		},
		{
			name:       "app route",
			path:       "/packs/123",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
		},
		{
			name:       "app route with HEAD",
			method:     http.MethodHead,
			path:       "/packs/123",
			wantStatus: http.StatusOK,
		},
		{
			name:       "directory",
			path:       "/_app/immutable",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
		},
		{
			name:       "directory with trailing slash",
			path:       "/_app/",
			wantStatus: http.StatusOK,
			wantBody:   indexHTML,
		},
		{
			name:       "directory whose name has an extension",
			path:       "/_app/immutable/assets/dir.css",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "immutable file",
			path:       "/_app/immutable/entry/app.123.js",
			wantStatus: http.StatusOK,
			wantBody:   appJS,
			wantType:   "text/javascript; charset=utf-8",
			wantCache:  immutable,
		},
		{
			name:       "other file in _app",
			path:       "/_app/version.json",
			wantStatus: http.StatusOK,
			wantBody:   `{"version":"1"}`,
		},
		{
			name:       "file at the root",
			path:       "/robots.txt",
			wantStatus: http.StatusOK,
			wantBody:   robotsTXT,
		},
		{
			name:       "missing immutable file",
			path:       "/_app/immutable/entry/app.old.js",
			wantStatus: http.StatusNotFound,
		},
		{
			// ServeMux redirects to the cleaned path, so the request never
			// reaches the front end's handler with ".." in it.
			name:       "path that leaves the root",
			path:       "/../robots.txt",
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:       "front end with wrong method",
			method:     http.MethodPost,
			path:       "/packs/123",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "API",
			path:       "/api/unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "API with another method",
			method:     http.MethodPost,
			path:       "/api/unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "API root",
			path:       "/api/",
			wantStatus: http.StatusNotFound,
		},
	}

	handler := server.New(newFrontend())

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), test.method, test.path, http.NoBody)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, test.wantStatus, rec.Code)
			assert.Equal(t, test.wantCache, rec.Header().Get("Cache-Control"), "Cache-Control")

			if test.wantBody != "" {
				assert.Equal(t, test.wantBody, rec.Body.String())
			}

			if test.wantType != "" {
				assert.Equal(t, test.wantType, rec.Header().Get("Content-Type"))
			}
		})
	}
}

func TestFrontendAllowHeader(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", http.NoBody)
	rec := httptest.NewRecorder()
	server.New(newFrontend()).ServeHTTP(rec, req)

	assert.Equal(t, "GET, HEAD", rec.Header().Get("Allow"))
}
