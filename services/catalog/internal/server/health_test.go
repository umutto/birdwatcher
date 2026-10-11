package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/server"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.err
}

func TestServer_HandleLivez(t *testing.T) {
	t.Run("returns 200 ok with valid status json", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, nil, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if body.Status != "ok" {
			t.Errorf("expected status ok, got %q", body.Status)
		}
	})
}

func TestServer_HandleReadyz(t *testing.T) {
	t.Run("returns 200 ok when database is healthy", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), &mockPinger{err: nil}, nil, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if body.Status != "ok" {
			t.Errorf("expected status ok, got %q", body.Status)
		}
	})

	t.Run("returns 503 service unavailable when database fails", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), &mockPinger{err: errors.New("connection failed")}, nil, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, rec.Code)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var body struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if body.Status != "unavailable" {
			t.Errorf("expected status unavailable, got %q", body.Status)
		}
	})
}
