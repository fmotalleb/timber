package response

import "net/http"

// PermissionDenied writes a permission denied response to the client.
func PermissionDenied(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="logs"`)
	http.Error(w, "Permission denied", http.StatusForbidden)
}
