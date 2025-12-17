package response

import "net/http"

// Unauthorized writes an unauthorized response to the client.
func Unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="logs"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
