package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/services/catalog/internal/server"
)

type mockSpeciesRepo struct {
	repository.Querier
	species    []repository.Species
	dictionary []repository.ListSpeciesDictionaryRow
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

func assertErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantError string) {
	t.Helper()

	if rec.Code != wantStatus {
		t.Fatalf("got status %d, want %d", rec.Code, wantStatus)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("got Content-Type %q, want %q", ct, "application/json")
	}

	var got struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if got.Error != wantError {
		t.Errorf("got error %q, want %q", got.Error, wantError)
	}
}

func TestServer_HandleListSpecies(t *testing.T) {
	t.Run("returns 200 with a json list of all species", func(t *testing.T) {
		want := []repository.Species{
			{
				PlateNumber:    1,
				CommonName:     "commonName",
				Aliases:        []string{"alias1", "alias2"},
				ScientificName: "scientificName",
				ImageUrl:       "imageURL",
				Metadata:       json.RawMessage(`{}`),
			},
		}
		mockRepo := &mockSpeciesRepo{
			species: want,
		}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("got Content-Type %q, want %q", ct, "application/json")
		}

		var got []repository.Species
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got species %+v, want %+v", got, want)
		}
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mockRepo := &mockSpeciesRepo{err: errors.New("db error")}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleListSpeciesDictionary(t *testing.T) {
	t.Run("returns 200 with a json list of all species name attributes", func(t *testing.T) {
		want := []repository.ListSpeciesDictionaryRow{
			{
				PlateNumber:    1,
				CommonName:     "commonName",
				Aliases:        []string{"alias1", "alias2"},
				ScientificName: "scientificName",
			},
		}
		mockRepo := &mockSpeciesRepo{
			dictionary: want,
		}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/dictionary", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("got Content-Type %q, want %q", ct, "application/json")
		}

		var got []repository.ListSpeciesDictionaryRow
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got dictionary %+v, want %+v", got, want)
		}
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		mockRepo := &mockSpeciesRepo{err: errors.New("db error")}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/dictionary", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}

func TestServer_HandleGetSpecies(t *testing.T) {
	t.Run("returns 200 with the species details of a plate number", func(t *testing.T) {
		want := repository.Species{
			PlateNumber:    1,
			CommonName:     "commonName",
			Aliases:        []string{"alias1", "alias2"},
			ScientificName: "scientificName",
			ImageUrl:       "imageURL",
			Metadata:       json.RawMessage(`{}`),
		}
		mockRepo := &mockSpeciesRepo{
			species: []repository.Species{want},
		}
		s := server.New(slog.New(slog.DiscardHandler), nil, mockRepo, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/1", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("got Content-Type %q, want %q", ct, "application/json")
		}

		var got repository.Species
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got species %+v, want %+v", got, want)
		}
	})

	t.Run("returns 400 when plate number is invalid", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, &mockSpeciesRepo{}, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/INVALID", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusBadRequest, "invalid plate number")
	})

	t.Run("returns 404 when species not found", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, &mockSpeciesRepo{}, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/999", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusNotFound, "species not found")
	})

	t.Run("returns 500 when repository fails", func(t *testing.T) {
		s := server.New(
			slog.New(slog.DiscardHandler),
			nil,
			&mockSpeciesRepo{err: errors.New("db error")},
			config.ServerConfig{Port: 8080},
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/species/1", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		assertErrorResponse(t, rec, http.StatusInternalServerError, "internal server error")
	})
}
