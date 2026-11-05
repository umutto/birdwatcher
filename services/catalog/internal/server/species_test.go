package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/services/catalog/internal/server"
)

type mockSpeciesRepo struct {
	repository.Querier
	species    []repository.Species
	dictionary []repository.ListSpeciesDictionaryRow
	sightings  []repository.Sighting
	err        error
}

func (m *mockSpeciesRepo) ListAllSpecies(_ context.Context) ([]repository.Species, error) {
	return m.species, m.err
}

func (m *mockSpeciesRepo) ListSpeciesDictionary(_ context.Context) ([]repository.ListSpeciesDictionaryRow, error) {
	return m.dictionary, m.err
}

func (m *mockSpeciesRepo) GetSpeciesByPlateNumber(_ context.Context, plateNumber int32) (repository.Species, error) {
	if m.err != nil {
		return repository.Species{}, m.err
	}
	for _, s := range m.species {
		if s.PlateNumber == plateNumber {
			return s, nil
		}
	}
	return repository.Species{}, pgx.ErrNoRows
}

func (m *mockSpeciesRepo) ListSightingsBySpecies(_ context.Context, _ repository.ListSightingsBySpeciesParams) ([]repository.Sighting, error) {
	return m.sightings, m.err
}

func TestServer_HandleListSpecies(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with a json list of all species", func(t *testing.T) {
		t.Parallel()

		want := []repository.Species{
			{
				PlateNumber:    11,
				CommonName:     "Bald Eagle",
				Aliases:        []string{"bird of washington", "bald eagle"},
				ScientificName: "Haliaeetus leucocephalus",
				ImageUrl:       "images/plate-11.webp",
				Metadata:       json.RawMessage(`{}`),
			},
		}
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{species: want},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species", nil)
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
			&mockSpeciesRepo{err: errors.New("db error")},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleListSpeciesDictionary(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with a json list of all species name attributes", func(t *testing.T) {
		t.Parallel()

		want := []repository.ListSpeciesDictionaryRow{
			{
				PlateNumber:    11,
				CommonName:     "Bald Eagle",
				Aliases:        []string{"bird of washington", "bald eagle"},
				ScientificName: "Haliaeetus leucocephalus",
			},
		}
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{dictionary: want},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/dictionary", nil)
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
			&mockSpeciesRepo{err: errors.New("db error")},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/dictionary", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleGetSpecies(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with the species details of a plate number", func(t *testing.T) {
		t.Parallel()

		want := repository.Species{
			PlateNumber:    11,
			CommonName:     "Bald Eagle",
			Aliases:        []string{"bird of washington", "bald eagle"},
			ScientificName: "Haliaeetus leucocephalus",
			ImageUrl:       "images/plate-11.webp",
			Metadata:       json.RawMessage(`{}`),
		}
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{species: []repository.Species{want}},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/11", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, want)
	})

	t.Run("returns 400 when plate number is invalid", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/INVALID", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid plate number")
	})

	t.Run("returns 404 when species not found", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/999", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusNotFound, "species not found")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{err: errors.New("db error")},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/11", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleListSpeciesSightings(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with a json list of sightings for a plate number", func(t *testing.T) {
		t.Parallel()

		snippet := "saw a Bald Eagle perched near the river"
		want := []repository.Sighting{
			{
				ID:                 uuid.MustParse("0195c000-0000-7000-8000-000000000001"),
				CrawlID:            uuid.MustParse("0195c000-0000-7000-8000-000000000002"),
				SpeciesPlateNumber: 11,
				SourceType:         repository.SightingSourceTypeText,
				Confidence:         0.95,
				PageUrl:            "https://example.com/birds",
				Snippet:            &snippet,
				Metadata:           json.RawMessage(`{}`),
				CreatedAt:          time.Date(2026, time.October, 24, 12, 0, 0, 0, time.UTC),
			},
		}
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{sightings: want},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/11/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertJSONResponse(t, rec, http.StatusOK, want)
	})

	t.Run("returns 400 when plate number is invalid", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/INVALID/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid plate number")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		t.Parallel()

		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			&mockSpeciesRepo{err: errors.New("db error")},
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/11/sightings", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}
