package server

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

// handleListSpecies returns the full master taxonomy of Audubon bird species.
func (s *Server) handleListSpecies(w http.ResponseWriter, r *http.Request) {
	species, err := s.repo.ListAllSpecies(r.Context())
	if err != nil {
		s.log.ErrorContext(r.Context(), "list all species", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, species)
}

// handleListSpeciesDictionary returns the lightweight species dictionary for text matching.
func (s *Server) handleListSpeciesDictionary(w http.ResponseWriter, r *http.Request) {
	dictionary, err := s.repo.ListSpeciesDictionary(r.Context())
	if err != nil {
		s.log.ErrorContext(r.Context(), "list species dictionary", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, dictionary)
}

// handleGetSpecies returns the details of a species for a given plate number.
func (s *Server) handleGetSpecies(w http.ResponseWriter, r *http.Request) {
	plate, err := strconv.ParseInt(r.PathValue("plate"), 10, 32)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid plate number")
		return
	}

	species, err := s.repo.GetSpeciesByPlateNumber(r.Context(), int32(plate))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "species not found")
		return
	}
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"get species by plate number",
			slog.Int64("plate_number", plate),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, species)
}
