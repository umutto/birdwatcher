package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/services/catalog/internal/server"
)

var (
	crawlID   = uuid.MustParse("0195c000-0000-7000-8000-000000000001")
	mockCrawl = repository.Crawl{
		ID:        crawlID,
		TargetUrl: "https://example.com/birds",
		State:     repository.CrawlStatePending,
	}
)

type mockCrawlsRepo struct {
	repository.Querier
	crawls         []repository.Crawl
	crawlSummaries []repository.ListCrawlsWithSummaryRow
	catalog        []repository.GetCatalogForCrawlRow
	sightings      []repository.ListSightingsByCrawlRow
	err            error
}

func (m *mockCrawlsRepo) CreateCrawl(_ context.Context, _ string) (repository.Crawl, error) {
	if m.err != nil {
		return repository.Crawl{}, m.err
	}
	return m.crawls[0], nil
}

func (m *mockCrawlsRepo) ListCrawlsWithSummary(_ context.Context, _ repository.ListCrawlsWithSummaryParams) ([]repository.ListCrawlsWithSummaryRow, error) {
	return m.crawlSummaries, m.err
}

func (m *mockCrawlsRepo) GetCrawlByID(_ context.Context, id uuid.UUID) (repository.Crawl, error) {
	if m.err != nil {
		return repository.Crawl{}, m.err
	}
	for _, c := range m.crawls {
		if c.ID == id {
			return c, nil
		}
	}
	return repository.Crawl{}, pgx.ErrNoRows
}

func (m *mockCrawlsRepo) GetCatalogForCrawl(_ context.Context, _ uuid.UUID) ([]repository.GetCatalogForCrawlRow, error) {
	return m.catalog, m.err
}

func (m *mockCrawlsRepo) ListSightingsByCrawl(_ context.Context, _ repository.ListSightingsByCrawlParams) ([]repository.ListSightingsByCrawlRow, error) {
	return m.sightings, m.err
}

func TestServer_HandleCreateCrawl(t *testing.T) {
	t.Run("returns 201 with the created crawl", func(t *testing.T) {
		mockRepo := &mockCrawlsRepo{crawls: []repository.Crawl{mockCrawl}}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		body := strings.NewReader(`{"target_url": "https://example.com/birds"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", body)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusCreated, mockCrawl)
	})

	t.Run("returns 400 when request body is invalid", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, &mockCrawlsRepo{}, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", strings.NewReader(`INVALID`))
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid request body")
	})

	t.Run("returns 400 when target url is invalid", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, &mockCrawlsRepo{}, config.ServerConfig{Port: 8080})

		body := strings.NewReader(`{"target_url": "INVALID"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", body)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid target url")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mockRepo := &mockCrawlsRepo{err: errors.New("db error")}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		body := strings.NewReader(`{"target_url": "https://example.com/birds"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", body)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleListCrawls(t *testing.T) {
	t.Run("returns 200 with a json list of crawls", func(t *testing.T) {
		want := []repository.ListCrawlsWithSummaryRow{
			{
				ID:                 crawlID,
				TargetUrl:          "https://example.com/birds",
				State:              repository.CrawlStateCompleted,
				TotalSightings:     42,
				UniqueSpeciesCount: 7,
			},
		}
		mockRepo := &mockCrawlsRepo{crawlSummaries: want}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, want)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mockRepo := &mockCrawlsRepo{err: errors.New("db error")}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleGetCrawl(t *testing.T) {
	t.Run("returns 200 with the crawl details of a crawl id", func(t *testing.T) {
		mockRepo := &mockCrawlsRepo{crawls: []repository.Crawl{mockCrawl}}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String(), nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, mockCrawl)
	})

	t.Run("returns 400 when crawl id is invalid", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, &mockCrawlsRepo{}, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/INVALID", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid crawl id")
	})

	t.Run("returns 404 when crawl not found", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, &mockCrawlsRepo{}, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String(), nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusNotFound, "crawl not found")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mockRepo := &mockCrawlsRepo{err: errors.New("db error")}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String(), nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleGetCrawlCatalog(t *testing.T) {
	t.Run("returns 200 with a json list of catalog species for a crawl id", func(t *testing.T) {
		want := []repository.GetCatalogForCrawlRow{
			{
				PlateNumber:    11,
				CommonName:     "Bald Eagle",
				Aliases:        []string{"bird of washington", "bald eagle"},
				ScientificName: "Haliaeetus leucocephalus",
				ImageUrl:       "images/plate-11.webp",
				Metadata:       json.RawMessage(`{}`),
				SightingCount:  3,
			},
		}
		mockRepo := &mockCrawlsRepo{catalog: want}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String()+"/catalog", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, want)
	})

	t.Run("returns 400 when crawl id is invalid", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, &mockCrawlsRepo{}, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/INVALID/catalog", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid crawl id")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mockRepo := &mockCrawlsRepo{err: errors.New("db error")}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String()+"/catalog", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleListCrawlSightings(t *testing.T) {
	t.Run("returns 200 with a json list of sightings for a crawl id", func(t *testing.T) {
		want := []repository.ListSightingsByCrawlRow{
			{
				ID:                 uuid.MustParse("0195c000-0000-7000-8000-000000000002"),
				CrawlID:            crawlID,
				SpeciesPlateNumber: 11,
				SourceType:         repository.SightingSourceTypeText,
				Confidence:         0.95,
				PageUrl:            "https://example.com/birds",
				Metadata:           json.RawMessage(`{}`),
				CommonName:         "Bald Eagle",
				ScientificName:     "Haliaeetus leucocephalus",
			},
		}
		mockRepo := &mockCrawlsRepo{sightings: want}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String()+"/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, want)
	})

	t.Run("returns 400 when crawl id is invalid", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, &mockCrawlsRepo{}, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/INVALID/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid crawl id")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mockRepo := &mockCrawlsRepo{err: errors.New("db error")}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String()+"/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}
