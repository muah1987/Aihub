package httputil

import (
	"encoding/json"
	"net/http"
)

// MaxBodySize is the default maximum request body size (1 MB).
const MaxBodySize = 1 << 20

// LargeBodySize is the maximum for endpoints that accept larger content (10 MB).
const LargeBodySize = 10 << 20

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// WriteError writes a JSON error response.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// LimitBody limits the request body size and returns the limited reader.
func LimitBody(w http.ResponseWriter, r *http.Request, maxBytes int64) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
}
