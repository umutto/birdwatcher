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

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

func TestServer_HandleLivez(t *testing.T) {
	t.Run("returns 200 ok with valid status json", func(t *testing.T) {
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			nil,
			nil,
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/livez", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("got Content-Type %q, want %q", ct, "application/json")
		}

		var got struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if got.Status != "ok" {
			t.Errorf("got status %q, want %q", got.Status, "ok")
		}
	})
}

func TestServer_HandleReadyz(t *testing.T) {
	t.Run("returns 200 ok when database is healthy", func(t *testing.T) {
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			&mockPinger{err: nil},
			nil,
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("got Content-Type %q, want %q", ct, "application/json")
		}

		var got struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if got.Status != "ok" {
			t.Errorf("got status %q, want %q", got.Status, "ok")
		}
	})

	t.Run("returns 503 service unavailable when database fails", func(t *testing.T) {
		s := server.New(
			slog.New(slog.DiscardHandler),
			config.ServerConfig{Port: 8080},
			&mockPinger{err: errors.New("connection failed")},
			nil,
			nil,
		)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("got status %d, want %d", rec.Code, http.StatusServiceUnavailable)
		}

		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("got Content-Type %q, want %q", ct, "application/json")
		}

		var got struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to decode response body: %v", err)
		}
		if got.Status != "unavailable" {
			t.Errorf("got status %q, want %q", got.Status, "unavailable")
		}
	})
}
