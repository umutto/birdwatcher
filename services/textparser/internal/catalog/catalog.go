// Package catalog provides an HTTP client for fetching species dictionary data from the catalog service.
package catalog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/umutto/birdwatcher/services/textparser/internal/config"
)

const dictionaryPath = "/api/v1/species/dictionary"

// Species represents a bird species dictionary entry returned by the catalog API.
type Species struct {
	PlateNumber    int32    `json:"plate_number"`
	CommonName     string   `json:"common_name"`
	ScientificName string   `json:"scientific_name"`
	Aliases        []string `json:"aliases"`
}

// Client retrieves species data from the catalog HTTP API.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// New creates an initialized Client with the provided catalog configuration.
func New(cfg config.CatalogConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		baseURL: cfg.URL,
	}
}

// FetchDictionary retrieves the species dictionary from the catalog service.
func (c *Client) FetchDictionary(ctx context.Context) ([]Species, error) {
	reqURL, err := url.JoinPath(c.baseURL, dictionaryPath)
	if err != nil {
		return nil, fmt.Errorf("join path: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var species []Species
	if err := json.NewDecoder(resp.Body).Decode(&species); err != nil {
		return nil, fmt.Errorf("decode dictionary: %w", err)
	}
	if len(species) == 0 {
		return nil, errors.New("empty catalog dictionary")
	}

	return species, nil
}
