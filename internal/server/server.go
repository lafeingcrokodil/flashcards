// Package server defines the HTTP routes of the app.
package server

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// immutablePrefix is the directory of the front end whose files never change
// without also changing their names, so browsers may cache them indefinitely.
const immutablePrefix = "_app/immutable/"

// immutableCacheControl lets browsers keep a file for a year without checking
// whether it has changed.
const immutableCacheControl = "public, max-age=31536000, immutable"

// New returns a handler that serves all of the app's routes, plus the front
// end from the given files, which must include an index.html.
func New(frontend fs.FS) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)
	// Requests under /api/ never get the front end, so that API clients get an
	// error rather than an HTML page for routes that don't exist.
	mux.Handle("/api/", http.NotFoundHandler())
	// This can't be "GET /", because ServeMux would reject it as conflicting
	// with "/api/", so handleFrontend checks the method itself.
	mux.Handle("/", handleFrontend(frontend))

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

// handleFrontend serves a single-page app: requests for existing files get
// those files, while any other path without a file extension gets index.html,
// so that the app can handle the route in the browser. Requests for missing
// files with an extension (e.g. an outdated script) get a 404 instead of HTML.
func handleFrontend(files fs.FS) http.Handler {
	fileServer := http.FileServerFS(files)

	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			resp.Header().Set("Allow", "GET, HEAD")
			http.Error(resp, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)

			return
		}

		name := strings.TrimPrefix(path.Clean(req.URL.Path), "/")

		info, err := fs.Stat(files, name)
		if err == nil && !info.IsDir() {
			if strings.HasPrefix(name, immutablePrefix) {
				resp.Header().Set("Cache-Control", immutableCacheControl)
			}

			fileServer.ServeHTTP(resp, req)

			return
		}

		if path.Ext(name) != "" {
			http.NotFound(resp, req)

			return
		}

		http.ServeFileFS(resp, req, files, "index.html")
	})
}
