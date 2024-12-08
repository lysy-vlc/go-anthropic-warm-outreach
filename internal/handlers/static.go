package handlers

import (
	"net/http"
	"path/filepath"
)

func (h *Handlers) HandleStatic(root string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(root))
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the file path from the URL
		filepath.Clean(r.URL.Path)

		// Serve the file
		http.StripPrefix("/static/", fs).ServeHTTP(w, r)
	}
}
