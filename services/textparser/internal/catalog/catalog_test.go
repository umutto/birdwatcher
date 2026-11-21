package catalog_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/textparser/internal/catalog"
	"github.com/umutto/birdwatcher/services/textparser/internal/config"
)

func TestClient_FetchDictionary(t *testing.T) {
	t.Parallel()

	t.Run("fetches and decodes species dictionary", func(t *testing.T) {
		t.Parallel()

		want := []catalog.Species{
			{
				PlateNumber:    11,
				CommonName:     "Bald Eagle",
				ScientificName: "Haliaeetus leucocephalus",
				Aliases:        []string{"bird of washington", "bald eagle"},
			},
		}

		var gotMethod string
		var gotPath string
		var gotAccept string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			gotAccept = r.Header.Get("Accept")

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(want)
		}))
		defer server.Close()

		client := catalog.New(config.CatalogConfig{
			URL:     server.URL + "/",
			Timeout: 2 * time.Second,
		})

		got, err := client.FetchDictionary(context.Background())
		if err != nil {
			t.Fatalf("FetchDictionary() unexpected error: %v", err)
		}

		if gotMethod != http.MethodGet {
			t.Errorf("request method = %q, want %q", gotMethod, http.MethodGet)
		}
		if gotPath != "/api/v1/species/dictionary" {
			t.Errorf("request path = %q, want %q", gotPath, "/api/v1/species/dictionary")
		}
		if gotAccept != "application/json" {
			t.Errorf("Accept header = %q, want %q", gotAccept, "application/json")
		}

		equalSpecies := func(a, b catalog.Species) bool {
			return a.PlateNumber == b.PlateNumber &&
				a.CommonName == b.CommonName &&
				a.ScientificName == b.ScientificName &&
				slices.Equal(a.Aliases, b.Aliases)
		}
		if !slices.EqualFunc(got, want, equalSpecies) {
			t.Errorf("FetchDictionary() = %+v, want %+v", got, want)
		}
	})

	t.Run("rejects non-200 status codes", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "internal error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := catalog.New(config.CatalogConfig{
			URL:     server.URL,
			Timeout: 2 * time.Second,
		})

		if _, err := client.FetchDictionary(context.Background()); err == nil {
			t.Fatal("FetchDictionary() expected error for 500 status, got nil")
		}
	})

	t.Run("rejects invalid json", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{invalid"))
		}))
		defer server.Close()

		client := catalog.New(config.CatalogConfig{
			URL:     server.URL,
			Timeout: 2 * time.Second,
		})

		if _, err := client.FetchDictionary(context.Background()); err == nil {
			t.Fatal("FetchDictionary() expected error for invalid json, got nil")
		}
	})

	t.Run("rejects empty dictionary", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("[]"))
		}))
		defer server.Close()

		client := catalog.New(config.CatalogConfig{
			URL:     server.URL,
			Timeout: 2 * time.Second,
		})

		if _, err := client.FetchDictionary(context.Background()); err == nil {
			t.Fatal("FetchDictionary() expected error for empty dictionary, got nil")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		}))
		defer server.Close()

		client := catalog.New(config.CatalogConfig{
			URL:     server.URL,
			Timeout: 2 * time.Second,
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		if _, err := client.FetchDictionary(ctx); err == nil {
			t.Fatal("FetchDictionary() expected context error, got nil")
		}
	})
}
