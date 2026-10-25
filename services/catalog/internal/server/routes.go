package server

import "net/http"

// routes registers application routes and returns the root http.Handler.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// probe handlers
	mux.HandleFunc("GET /livez", s.handleLivez)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
	mux.HandleFunc("GET /healthz", s.handleLivez)

	// species handlers
	mux.HandleFunc("GET /api/v1/species", s.handleListSpecies)
	mux.HandleFunc("GET /api/v1/species/dictionary", s.handleListSpeciesDictionary)
	mux.HandleFunc("GET /api/v1/species/{plate}", s.handleGetSpecies)
	mux.HandleFunc("GET /api/v1/species/{plate}/sightings", s.handleListSpeciesSightings)

	// crawls handlers
	mux.HandleFunc("POST /api/v1/crawls", s.handleCreateCrawl)
	mux.HandleFunc("GET /api/v1/crawls", s.handleListCrawls)
	mux.HandleFunc("GET /api/v1/crawls/{id}", s.handleGetCrawl)
	mux.HandleFunc("GET /api/v1/crawls/{id}/catalog", s.handleGetCrawlCatalog)
	mux.HandleFunc("GET /api/v1/crawls/{id}/sightings", s.handleListCrawlSightings)

	return chain(
		mux,
		requestLogger(s.log),
	)
}
