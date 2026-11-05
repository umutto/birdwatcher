// Package config manages configuration loading and validation for the frontier service.
package config

import (
	"errors"
	"strings"
	"time"

	"github.com/umutto/birdwatcher/shared/env"
)

// ServerConfig holds gRPC server configuration options.
type ServerConfig struct {
	Port int32
}

// NATSConfig holds connection settings for NATS JetStream.
type NATSConfig struct {
	URL string
}

// ValkeyConfig holds connection parameters for the Valkey instance.
type ValkeyConfig struct {
	Addr string
}

// CrawlConfig holds scheduling, rate limiting and politeness parameters.
type CrawlConfig struct {
	HostDelay          time.Duration
	MaxInflightPerHost int32
}

// Config aggregates all runtime configuration for the frontier service.
type Config struct {
	Env    env.Environment
	Server ServerConfig
	NATS   NATSConfig
	Valkey ValkeyConfig
	Crawl  CrawlConfig
}

// Load reads environment variables and returns a validated Config.
func Load() (*Config, error) {
	var errs []error

	port, err := env.Int32("PORT", 9090)
	errs = append(errs, err)

	natsURL := env.String("NATS_URL", "nats://localhost:4222")

	valkeyAddr := env.String("VALKEY_ADDR", "localhost:6379")

	hostDelay, err := env.Duration("CRAWL_HOST_DELAY", 250*time.Millisecond)
	errs = append(errs, err)

	maxInflightPerHost, err := env.Int32("CRAWL_MAX_INFLIGHT_PER_HOST", 2)
	errs = append(errs, err)

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	cfg := &Config{
		Env: env.Current(),
		Server: ServerConfig{
			Port: port,
		},
		NATS: NATSConfig{
			URL: natsURL,
		},
		Valkey: ValkeyConfig{
			Addr: valkeyAddr,
		},
		Crawl: CrawlConfig{
			HostDelay:          hostDelay,
			MaxInflightPerHost: maxInflightPerHost,
		},
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate verifies logical invariants that downstream drivers wouldn't catch cleanly.
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return errors.New("server port must be between 1 and 65535")
	}
	if strings.TrimSpace(c.NATS.URL) == "" {
		return errors.New("nats url cannot be empty")
	}
	if strings.TrimSpace(c.Valkey.Addr) == "" {
		return errors.New("valkey addr cannot be empty")
	}
	if c.Crawl.HostDelay < 0 {
		return errors.New("host delay cannot be negative")
	}
	if c.Crawl.MaxInflightPerHost < 1 {
		return errors.New("max inflight per host cannot be less than 1")
	}
	return nil
}
