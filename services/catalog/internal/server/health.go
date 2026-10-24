package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// healthResponse defines the JSON schema for the health check endpoint.
type healthResponse struct {
	Status string `json:"status"`
}

// handleLivez verifies the process is responsive.
func (s *Server) handleLivez(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// handleReadyz verifies the database connection is healthy before serving traffic.
func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.db.Ping(ctx); err != nil {
		s.log.ErrorContext(r.Context(), "readiness probe failed: database ping", slog.Any("error", err))
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
