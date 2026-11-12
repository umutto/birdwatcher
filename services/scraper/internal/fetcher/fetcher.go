// Package fetcher provides an HTTP client for retrieving web pages with timeouts,
// size bounds, and content-type verification.
package fetcher

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/umutto/birdwatcher/services/scraper/internal/config"
)

const userAgent = "Birdwatcher/1.0 (+https://github.com/umutto/birdwatcher)"

// Fetcher retrieves HTML pages over HTTP.
type Fetcher struct {
	client       *http.Client
	maxBodyBytes int32
}

// New creates an initialized Fetcher with the provided scraper configuration.
func New(cfg config.ScrapeConfig) *Fetcher {
	return &Fetcher{
		client:       &http.Client{Timeout: cfg.Timeout},
		maxBodyBytes: cfg.MaxBodyBytes,
	}
}

// Fetch downloads the HTML document at targetURL, enforcing timeouts, size limits,
// and verifying that the response is an HTML document.
func (f *Fetcher) Fetch(ctx context.Context, targetURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if mediaType != "text/html" && mediaType != "application/xhtml+xml" {
		return nil, fmt.Errorf("unsupported content type: %q", mediaType)
	}

	limitedReader := io.LimitReader(resp.Body, int64(f.maxBodyBytes))
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return body, nil
}
