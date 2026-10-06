package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/database"
)

func TestNew(t *testing.T) {
	t.Run("invalid connection string", func(t *testing.T) {
		cfg := validConfig()
		cfg.URL = "://invalid_url"

		pool, err := database.New(context.Background(), cfg)
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
		if pool != nil {
			t.Errorf("expected nil pool, got %+v", pool)
		}
	})

	t.Run("unreachable host", func(t *testing.T) {
		cfg := validConfig()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		pool, err := database.New(ctx, cfg)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected error wrapping context.Canceled, got %v", err)
		}
		if pool != nil {
			t.Errorf("expected nil pool, got %+v", pool)
		}
	})
}

func validConfig() config.DatabaseConfig {
	return config.DatabaseConfig{
		URL:               "postgres://username:password@database-url:5432/birdwatcher_dev?sslmode=disable",
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}
}
