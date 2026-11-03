// Package config manages configuration loading and validation for the catalog service.
package config

import (
	"errors"
	"strings"
	"time"

	"github.com/umutto/birdwatcher/shared/env"
)

// ServerConfig holds HTTP server configuration options.
type ServerConfig struct {
	Port     int32
	Frontier FrontierConfig
}

// FrontierConfig holds connection settings and default crawl limits for the Frontier service.
type FrontierConfig struct {
	Addr     string
	MaxDepth int32
	MaxPages int32
}

// NATSConfig holds connection settings for NATS JetStream.
type NATSConfig struct {
	URL string
}

// DatabaseConfig holds connection pool settings for PostgreSQL.
type DatabaseConfig struct {
	URL               string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

// Config aggregates all runtime configuration for the catalog service.
type Config struct {
	Env      env.Environment
	Server   ServerConfig
	NATS     NATSConfig
	Database DatabaseConfig
}

// Load reads environment variables and returns a validated Config.
func Load() (*Config, error) {
	var errs []error

	port, err := env.Int32("PORT", 8080)
	errs = append(errs, err)

	frontierAddr := env.String("FRONTIER_GRPC_ADDR", "localhost:9090")

	frontierMaxDepth, err := env.Int32("FRONTIER_MAX_DEPTH", 4)
	errs = append(errs, err)

	frontierMaxPages, err := env.Int32("FRONTIER_MAX_PAGES", 200)
	errs = append(errs, err)

	natsURL := env.String("NATS_URL", "nats://localhost:4222")

	dbURL, err := env.Require("DATABASE_URL")
	errs = append(errs, err)

	dbMaxConns, err := env.Int32("DATABASE_MAX_CONNS", 10)
	errs = append(errs, err)

	dbMinConns, err := env.Int32("DATABASE_MIN_CONNS", 2)
	errs = append(errs, err)

	dbMaxConnLifetime, err := env.Duration("DATABASE_MAX_CONN_LIFETIME", 1*time.Hour)
	errs = append(errs, err)

	dbMaxConnIdleTime, err := env.Duration("DATABASE_MAX_CONN_IDLE_TIME", 30*time.Minute)
	errs = append(errs, err)

	dbHealthCheckPeriod, err := env.Duration("DATABASE_HEALTH_CHECK_PERIOD", 30*time.Second)
	errs = append(errs, err)

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	cfg := &Config{
		Env: env.Current(),
		Server: ServerConfig{
			Port: port,
			Frontier: FrontierConfig{
				Addr:     frontierAddr,
				MaxDepth: frontierMaxDepth,
				MaxPages: frontierMaxPages,
			},
		},
		NATS: NATSConfig{
			URL: natsURL,
		},
		Database: DatabaseConfig{
			URL:               dbURL,
			MaxConns:          dbMaxConns,
			MinConns:          dbMinConns,
			MaxConnLifetime:   dbMaxConnLifetime,
			MaxConnIdleTime:   dbMaxConnIdleTime,
			HealthCheckPeriod: dbHealthCheckPeriod,
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
	if strings.TrimSpace(c.Server.Frontier.Addr) == "" {
		return errors.New("frontier grpc address cannot be empty")
	}
	if c.Server.Frontier.MaxDepth < 0 {
		return errors.New("frontier max_depth cannot be negative")
	}
	if c.Server.Frontier.MaxPages <= 0 {
		return errors.New("frontier max_pages must be greater than 0")
	}
	if strings.TrimSpace(c.NATS.URL) == "" {
		return errors.New("nats url cannot be empty")
	}
	if c.Database.MinConns > c.Database.MaxConns {
		return errors.New("database min_conns cannot exceed max_conns")
	}
	return nil
}
