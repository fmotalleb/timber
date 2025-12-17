// Package response provides helper functions for writing HTTP responses.
package response

import (
	"encoding/json"
	"net/http"
)

// JSON writes a JSON response to the client.
func JSON(w http.ResponseWriter, data any, status int) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	_, err = w.Write(b)
	return err
}

