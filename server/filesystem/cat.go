// package filesystem contains fs logic
package filesystem

import (
	"net/http"

	"github.com/fmotalleb/timber/server/helper"
)

// Cat serves a file to the client.
func Cat(w http.ResponseWriter, r *http.Request) {
	filePath, ok := helper.GetPath(r)
	if !ok {
		http.Error(w, "file path is missing from request, your request must contain `path` query parameter", http.StatusBadRequest)
		return
	}
	if containsDotDot(filePath) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filePath)
}
