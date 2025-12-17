package helper

import "net/http"

func GetPath(r *http.Request) (string, bool) {
	queries := r.URL.Query()
	if !queries.Has("path") {
		return "", false
	}
	reqPath := queries.Get("path")
	return reqPath, true
}
