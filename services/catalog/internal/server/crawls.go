package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
)

// createCrawlRequest defines the expected JSON payload for starting a new crawl job.
type createCrawlRequest struct {
	TargetURL string `json:"target_url"`
}

// handleCreateCrawl validates the target URL, creates a new crawl job, and returns it with HTTP 201.
func (s *Server) handleCreateCrawl(w http.ResponseWriter, r *http.Request) {
	var req createCrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := url.Parse(strings.TrimSpace(req.TargetURL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		writeError(w, http.StatusBadRequest, "invalid target url")
		return
	}

	targetURL := u.String()
	crawl, err := s.repo.CreateCrawl(r.Context(), targetURL)
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"create crawl",
			slog.String("target_url", targetURL),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	jobRequest := &birdwatcherv1.StartJobRequest{
		JobId:    crawl.ID.String(),
		RootUrl:  targetURL,
		MaxDepth: s.cfg.Frontier.MaxDepth,
		MaxPages: s.cfg.Frontier.MaxPages,
	}

	rpcCtx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if _, err := s.frontier.StartJob(rpcCtx, jobRequest); err != nil {
		s.log.ErrorContext(
			r.Context(),
			"start frontier job",
			slog.String("crawl_id", jobRequest.JobId),
			slog.String("target_url", jobRequest.RootUrl),
			slog.Int("max_depth", int(jobRequest.MaxDepth)),
			slog.Int("max_pages", int(jobRequest.MaxPages)),
			slog.Any("error", err),
		)
		if _, dbErr := s.repo.UpdateCrawlState(
			r.Context(),
			repository.UpdateCrawlStateParams{
				ID:    crawl.ID,
				State: repository.CrawlStateFailed,
			},
		); dbErr != nil {
			s.log.ErrorContext(
				r.Context(),
				"update crawl state to failed",
				slog.String("crawl_id", crawl.ID.String()),
				slog.Any("error", dbErr),
			)
		}
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	crawl, err = s.repo.UpdateCrawlState(
		r.Context(),
		repository.UpdateCrawlStateParams{
			ID:    crawl.ID,
			State: repository.CrawlStateRunning,
		},
	)
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"update crawl state to running",
			slog.String("crawl_id", crawl.ID.String()),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, crawl)
}

// handleListCrawls returns a paginated list of crawl jobs with aggregated sighting summaries.
func (s *Server) handleListCrawls(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	crawls, err := s.repo.ListCrawlsWithSummary(
		r.Context(),
		repository.ListCrawlsWithSummaryParams{
			LimitRows:  limit,
			OffsetRows: offset,
		},
	)
	if err != nil {
		s.log.ErrorContext(r.Context(), "list crawls", slog.Any("error", err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, crawls)
}

// handleGetCrawl returns the status and progress counters of a single crawl job by its UUID.
func (s *Server) handleGetCrawl(w http.ResponseWriter, r *http.Request) {
	crawlID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid crawl id")
		return
	}

	crawl, err := s.repo.GetCrawlByID(r.Context(), crawlID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "crawl not found")
		return
	}
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"get crawl by id",
			slog.String("crawl_id", crawlID.String()),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, crawl)
}

// handleDeleteCrawl deletes a finished crawl job and its cascading sightings by UUID.
func (s *Server) handleDeleteCrawl(w http.ResponseWriter, r *http.Request) {
	crawlID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid crawl id")
		return
	}

	crawl, err := s.repo.GetCrawlByID(r.Context(), crawlID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "crawl not found")
		return
	}
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"get crawl by id",
			slog.String("crawl_id", crawlID.String()),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if !isTerminalCrawlState(crawl.State) {
		writeError(w, http.StatusConflict, "cannot delete an active crawl")
		return
	}

	rows, err := s.repo.DeleteCrawl(r.Context(), crawlID)
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"delete crawl",
			slog.String("crawl_id", crawlID.String()),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if rows == 0 {
		writeError(w, http.StatusNotFound, "crawl not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetCrawlCatalog returns all 435 Audubon species annotated with sighting counts for a given crawl UUID.
func (s *Server) handleGetCrawlCatalog(w http.ResponseWriter, r *http.Request) {
	crawlID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid crawl id")
		return
	}

	catalog, err := s.repo.GetCatalogForCrawl(r.Context(), crawlID)
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"get catalog for crawl",
			slog.String("crawl_id", crawlID.String()),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, catalog)
}

// handleListCrawlSightings returns a paginated list of individual bird sightings recorded during a crawl job.
func (s *Server) handleListCrawlSightings(w http.ResponseWriter, r *http.Request) {
	crawlID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid crawl id")
		return
	}

	if plateParam := r.URL.Query().Get("plate"); plateParam != "" {
		s.listCrawlSightingsByPlate(w, r, crawlID, plateParam)
		return
	}

	limit, offset := parsePagination(r)

	sightings, err := s.repo.ListSightingsByCrawl(
		r.Context(),
		repository.ListSightingsByCrawlParams{
			CrawlID:    crawlID,
			LimitRows:  limit,
			OffsetRows: offset,
		},
	)
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"list sightings by crawl",
			slog.String("crawl_id", crawlID.String()),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, sightings)
}

// listCrawlSightingsByPlate returns a paginated list of sightings for the given species plate during a crawl job.
func (s *Server) listCrawlSightingsByPlate(w http.ResponseWriter, r *http.Request, crawlID uuid.UUID, plateParam string) {
	plate, err := strconv.ParseInt(plateParam, 10, 32)
	if err != nil || plate <= 0 {
		writeError(w, http.StatusBadRequest, "invalid plate number")
		return
	}
	limit, offset := parsePagination(r)

	sightings, err := s.repo.ListSightingsByCrawlAndSpecies(
		r.Context(),
		repository.ListSightingsByCrawlAndSpeciesParams{
			CrawlID:            crawlID,
			SpeciesPlateNumber: int32(plate),
			LimitRows:          limit,
			OffsetRows:         offset,
		},
	)
	if err != nil {
		s.log.ErrorContext(
			r.Context(),
			"list sightings by crawl and species",
			slog.String("crawl_id", crawlID.String()),
			slog.Int64("plate_number", plate),
			slog.Any("error", err),
		)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, sightings)
}

// isTerminalCrawlState reports whether a crawl has reached a final state.
func isTerminalCrawlState(state repository.CrawlState) bool {
	switch state {
	case repository.CrawlStateCompleted,
		repository.CrawlStateFailed,
		repository.CrawlStateCancelled:
		return true
	default:
		return false
	}
}
