package fetcher_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/scraper/internal/config"
	"github.com/umutto/birdwatcher/services/scraper/internal/fetcher"
)

func TestFetcher_Fetch(t *testing.T) {
	t.Parallel()

	t.Run("successful fetch with headers verified", func(t *testing.T) {
		t.Parallel()

		const expectedBody = "<html><body><h1>Bald Eagle</h1></body></html>"
		var receivedUserAgent string
		var receivedAccept string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedUserAgent = r.Header.Get("User-Agent")
			receivedAccept = r.Header.Get("Accept")

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(expectedBody))
		}))
		defer server.Close()

		f := fetcher.New(config.ScrapeConfig{
			Timeout:      2 * time.Second,
			MaxBodyBytes: 1024 * 1024,
		})

		body, err := f.Fetch(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("Fetch() unexpected error: %v", err)
		}

		if string(body) != expectedBody {
			t.Errorf("Fetch() body = %q, want %q", string(body), expectedBody)
		}
		if !strings.HasPrefix(receivedUserAgent, "Birdwatcher") {
			t.Errorf("User-Agent header = %q, want prefix 'Birdwatcher'", receivedUserAgent)
		}
		if !strings.Contains(receivedAccept, "text/html") {
			t.Errorf("Accept header = %q, want contains 'text/html'", receivedAccept)
		}
	})

	t.Run("rejects non-2xx status codes", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		}))
		defer server.Close()

		f := fetcher.New(config.ScrapeConfig{
			Timeout:      2 * time.Second,
			MaxBodyBytes: 1024,
		})

		_, err := f.Fetch(context.Background(), server.URL)
		if err == nil {
			t.Fatal("Fetch() expected error for 404 status, got nil")
		}
	})

	t.Run("rejects unsupported content type", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("%PDF-1.4..."))
		}))
		defer server.Close()

		f := fetcher.New(config.ScrapeConfig{
			Timeout:      2 * time.Second,
			MaxBodyBytes: 1024,
		})

		_, err := f.Fetch(context.Background(), server.URL)
		if err == nil {
			t.Fatal("Fetch() expected error for application/pdf, got nil")
		}
	})

	t.Run("enforces max body bytes limit", func(t *testing.T) {
		t.Parallel()

		const maxBytes = 50
		largePayload := strings.Repeat("A", 200)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(largePayload))
		}))
		defer server.Close()

		f := fetcher.New(config.ScrapeConfig{
			Timeout:      2 * time.Second,
			MaxBodyBytes: maxBytes,
		})

		body, err := f.Fetch(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("Fetch() unexpected error: %v", err)
		}

		if len(body) != maxBytes {
			t.Errorf("Fetch() body length = %d, want capped at %d", len(body), maxBytes)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		f := fetcher.New(config.ScrapeConfig{
			Timeout:      2 * time.Second,
			MaxBodyBytes: 1024,
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := f.Fetch(ctx, server.URL)
		if err == nil {
			t.Fatal("Fetch() expected context error, got nil")
		}
	})
}
