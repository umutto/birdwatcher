package server

import (
	"encoding/json"
	"net/http"
)

// errorResponse defines the standard JSON error payload for API endpoints.
type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON serializes data as JSON and writes it to the response with the given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeError writes a standardized JSON error response with the given HTTP status code.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
