package response

import "net/http"

func PermissionDenied(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="logs"`)
	http.Error(w, "Permission denied", http.StatusForbidden)
}
