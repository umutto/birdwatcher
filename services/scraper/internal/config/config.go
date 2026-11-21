// Package config manages configuration loading and validation for the scraper service.
package config

import (
	"errors"
	"strings"
	"time"

	"github.com/umutto/birdwatcher/shared/env"
)

// NATSConfig holds connection settings for NATS JetStream.
type NATSConfig struct {
	URL string
}

// ScrapeConfig holds timeout and size limits for outbound HTTP fetches.
type ScrapeConfig struct {
	Timeout      time.Duration
	MaxBodyBytes int32
}

// Config aggregates all runtime configuration for the scraper service.
type Config struct {
	Env    env.Environment
	NATS   NATSConfig
	Scrape ScrapeConfig
}

// Load reads environment variables and returns a validated Config.
func Load() (*Config, error) {
	var errs []error

	natsURL := env.String("NATS_URL", "nats://localhost:4222")

	timeout, err := env.Duration("HTTP_TIMEOUT", 10*time.Second)
	errs = append(errs, err)

	maxBodyBytes, err := env.Int32("HTTP_MAX_BODY_BYTES", 5242880)
	errs = append(errs, err)

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	cfg := &Config{
		Env: env.Current(),
		NATS: NATSConfig{
			URL: natsURL,
		},
		Scrape: ScrapeConfig{
			Timeout:      timeout,
			MaxBodyBytes: maxBodyBytes,
		},
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate verifies logical invariants that downstream drivers wouldn't catch cleanly.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.NATS.URL) == "" {
		return errors.New("nats url cannot be empty")
	}
	if c.Scrape.Timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if c.Scrape.MaxBodyBytes <= 0 {
		return errors.New("max body bytes must be greater than 0")
	}
	return nil
}
