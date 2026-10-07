package server_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/server"
)

func TestServer_HandleHealthz(t *testing.T) {
	t.Run("returns 200 ok with valid status json", func(t *testing.T) {
		s := server.New(slog.New(slog.DiscardHandler), nil, config.ServerConfig{Port: 8080})

		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", contentType)
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
