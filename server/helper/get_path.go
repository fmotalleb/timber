// Package helper provides helper functions for the server.
package helper

import "net/http"

// GetPath returns the path from the request.
func GetPath(r *http.Request) (string, bool) {
	queries := r.URL.Query()
	if !queries.Has("path") {
		return "", false
	}
	reqPath := queries.Get("path")
	return reqPath, true
}
