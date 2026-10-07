package server

import "net/http"

// routes registers application routes and returns the root http.Handler.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", s.handleHealthz)

	return mux
}
