package server

import (
	"encoding/json"
	"net/http"
)

// healthResponse defines the JSON schema for the health check endpoint.
type healthResponse struct {
	Status string `json:"status"`
}

// handleHealthz responds with service health status for orchestrators and load balancers.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}
