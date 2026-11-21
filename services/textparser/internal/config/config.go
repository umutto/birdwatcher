// Package config manages configuration loading and validation for the textparser service.
package config

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/umutto/birdwatcher/shared/env"
)

// NATSConfig holds connection settings for NATS JetStream.
type NATSConfig struct {
	URL string
}

// CatalogConfig holds connection and timeout settings for the Catalog HTTP API.
type CatalogConfig struct {
	URL     string
	Timeout time.Duration
}

// Config aggregates all runtime configuration for the textparser service.
type Config struct {
	Env     env.Environment
	NATS    NATSConfig
	Catalog CatalogConfig
}

// Load reads environment variables and returns a validated Config.
func Load() (*Config, error) {
	var errs []error

	natsURL := env.String("NATS_URL", "nats://localhost:4222")

	catalogURL := env.String("CATALOG_URL", "http://localhost:8080")

	catalogTimeout, err := env.Duration("CATALOG_TIMEOUT", 10*time.Second)
	errs = append(errs, err)

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	cfg := &Config{
		Env: env.Current(),
		NATS: NATSConfig{
			URL: natsURL,
		},
		Catalog: CatalogConfig{
			URL:     catalogURL,
			Timeout: catalogTimeout,
		},
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate verifies logical invariants across loaded configuration values.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.NATS.URL) == "" {
		return errors.New("nats url cannot be empty")
	}

	u, err := url.Parse(c.Catalog.URL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return errors.New("catalog url must be a valid http or https url")
	}

	if c.Catalog.Timeout <= 0 {
		return errors.New("catalog timeout must be greater than 0")
	}

	return nil
}
