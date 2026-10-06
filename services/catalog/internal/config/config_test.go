package config_test

import (
	"errors"
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/shared/env"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(c *config.Config)
		wantErr bool
	}{
		{
			name:    "valid config",
			mutate:  func(c *config.Config) {},
			wantErr: false,
		},
		{
			name: "port out of bounds",
			mutate: func(c *config.Config) {
				c.Server.Port = 70000
			},
			wantErr: true,
		},
		{
			name: "min_conns exceeds max_conns",
			mutate: func(c *config.Config) {
				c.Database.MaxConns = 5
				c.Database.MinConns = 10
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig()
			tt.mutate(cfg)

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	const testURL = "postgres://username:password@database-url:5432/birdwatcher_dev?sslmode=disable"

	t.Run("missing DATABASE_URL returns error", func(t *testing.T) {
		t.Setenv("DATABASE_URL", "")

		_, err := config.Load()
		if !errors.Is(err, env.ErrMissing) {
			t.Fatalf("expected ErrMissing, got: %v", err)
		}
	})

	t.Run("valid environment loads cleanly", func(t *testing.T) {
		t.Setenv("DATABASE_URL", testURL)

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("unexpected Load() error: %v", err)
		}
		if cfg.Database.URL != testURL {
			t.Errorf("expected URL %q, got %q", testURL, cfg.Database.URL)
		}
	})
}

func validConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port: 8080,
		},
		Database: config.DatabaseConfig{
			URL:               "postgres://username:password@database-url:5432/birdwatcher_dev?sslmode=disable",
			MaxConns:          10,
			MinConns:          2,
			MaxConnLifetime:   1 * time.Hour,
			MaxConnIdleTime:   30 * time.Minute,
			HealthCheckPeriod: 30 * time.Second,
		},
	}
}
