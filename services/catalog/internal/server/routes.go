package server

import "net/http"

// routes registers application routes and returns the root http.Handler.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /livez", s.handleLivez)
	mux.HandleFunc("GET /readyz", s.handleReadyz)

	// backwards compatibility alias
	mux.HandleFunc("GET /healthz", s.handleLivez)

	return chain(
		mux,
		requestLogger(s.log),
	)
}
