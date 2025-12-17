package response

import "net/http"

func Unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="logs"`)
	http.Error(w, "unauthorized", http.StatusUnauthorized)
}
