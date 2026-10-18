package server

import "net/http"

// handleListSpecies returns the full master taxonomy of Audubon bird species.
func (s *Server) handleListSpecies(w http.ResponseWriter, r *http.Request) {
	// 1. Call s.repo.ListAllSpecies(r.Context())
	// 2. On error: log with s.log.ErrorContext(r.Context(), ...), return 500 JSON errorResponse{Error: "internal server error"}
	// 3. On success: return 200 OK with Content-Type: application/json and encode the species slice

}
