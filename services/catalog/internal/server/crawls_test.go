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
	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/services/catalog/internal/server"
	"google.golang.org/grpc"
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

func (m *mockCrawlsRepo) UpdateCrawlState(_ context.Context, arg repository.UpdateCrawlStateParams) (repository.Crawl, error) {
	if m.err != nil {
		return repository.Crawl{}, m.err
	}
	c := m.crawls[0]
	c.State = arg.State
	return c, nil
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

type mockFrontierClient struct {
	err error
}

func (m *mockFrontierClient) StartJob(_ context.Context, _ *birdwatcherv1.StartJobRequest, _ ...grpc.CallOption) (*birdwatcherv1.StartJobResponse, error) {
	return &birdwatcherv1.StartJobResponse{}, m.err
}

func TestServer_HandleCreateCrawl(t *testing.T) {
	t.Parallel()

	t.Run("returns 201 with the created crawl", func(t *testing.T) {
		t.Parallel()

		want := mockCrawl
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{crawls: []repository.Crawl{want}},
			&mockFrontierClient{},
		)

		body := strings.NewReader(`{"target_url": "https://example.com/birds"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", body)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		want.State = repository.CrawlStateRunning
		assertJSONResponse(t, rec, http.StatusCreated, want)
	})

	t.Run("returns 400 when request body is invalid", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{},
			nil,
		)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", strings.NewReader(`INVALID`))
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid request body")
	})

	t.Run("returns 400 when target url is invalid", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{},
			nil,
		)

		body := strings.NewReader(`{"target_url": "INVALID"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", body)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid target url")
	})

	t.Run("returns 500 when frontier fails", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{crawls: []repository.Crawl{mockCrawl}},
			&mockFrontierClient{err: errors.New("grpc error")},
		)

		body := strings.NewReader(`{"target_url": "https://example.com/birds"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", body)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{err: errors.New("db error")},
			nil,
		)

		body := strings.NewReader(`{"target_url": "https://example.com/birds"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/crawls", body)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleListCrawls(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with a json list of crawls", func(t *testing.T) {
		t.Parallel()

		want := []repository.ListCrawlsWithSummaryRow{
			{
				ID:                 crawlID,
				TargetUrl:          "https://example.com/birds",
				State:              repository.CrawlStateCompleted,
				TotalSightings:     42,
				UniqueSpeciesCount: 7,
			},
		}
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{crawlSummaries: want},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, want)
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{err: errors.New("db error")},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleGetCrawl(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with the crawl details of a crawl id", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{crawls: []repository.Crawl{mockCrawl}},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String(), nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, mockCrawl)
	})

	t.Run("returns 400 when crawl id is invalid", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/INVALID", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid crawl id")
	})

	t.Run("returns 404 when crawl not found", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String(), nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusNotFound, "crawl not found")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{err: errors.New("db error")},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String(), nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleGetCrawlCatalog(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with a json list of catalog species for a crawl id", func(t *testing.T) {
		t.Parallel()

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
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{catalog: want},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String()+"/catalog", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, want)
	})

	t.Run("returns 400 when crawl id is invalid", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/INVALID/catalog", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid crawl id")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{err: errors.New("db error")},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String()+"/catalog", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleListCrawlSightings(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with a json list of sightings for a crawl id", func(t *testing.T) {
		t.Parallel()

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
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{sightings: want},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String()+"/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, want)
	})

	t.Run("returns 400 when crawl id is invalid", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/INVALID/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid crawl id")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockCrawlsRepo{err: errors.New("db error")},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/crawls/"+crawlID.String()+"/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}
