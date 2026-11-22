// Package web provides the built front end, embedded into the Go binary.
package web

import (
	"embed"
	"fmt"
	"io/fs"
)

// The all: prefix includes files whose names start with "_" or ".", such as
// SvelteKit's _app directory, which go:embed would skip otherwise.
//
//go:embed all:build
var build embed.FS

// Build returns the files of the built front end, with index.html at the root.
func Build() (fs.FS, error) {
	files, err := fs.Sub(build, "build")
	if err != nil {
		return nil, fmt.Errorf("opening the build directory: %w", err)
	}

	return files, nil
}
