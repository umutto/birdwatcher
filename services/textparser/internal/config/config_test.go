package config_test

import (
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/textparser/internal/config"
)

func TestConfig_Validate(t *testing.T) {
	t.Parallel()

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
			name: "empty nats url",
			mutate: func(c *config.Config) {
				c.NATS.URL = " "
			},
			wantErr: true,
		},
		{
			name: "empty catalog url",
			mutate: func(c *config.Config) {
				c.Catalog.URL = " "
			},
			wantErr: true,
		},
		{
			name: "non-http catalog url",
			mutate: func(c *config.Config) {
				c.Catalog.URL = "ftp://example.com"
			},
			wantErr: true,
		},
		{
			name: "malformed catalog url",
			mutate: func(c *config.Config) {
				c.Catalog.URL = "://invalid"
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			mutate: func(c *config.Config) {
				c.Catalog.Timeout = -1
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := validConfig()
			tt.mutate(cfg)

			err := cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func validConfig() *config.Config {
	return &config.Config{
		NATS: config.NATSConfig{
			URL: "nats://localhost:4222",
		},
		Catalog: config.CatalogConfig{
			URL:     "http://localhost:8080",
			Timeout: 10 * time.Second,
		},
	}
}
