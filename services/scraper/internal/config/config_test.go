package config_test

import (
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/scraper/internal/config"
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
			name: "negative timeout",
			mutate: func(c *config.Config) {
				c.Scrape.Timeout = -1
			},
			wantErr: true,
		},
		{
			name: "negative max body bytes",
			mutate: func(c *config.Config) {
				c.Scrape.MaxBodyBytes = -1
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
		Scrape: config.ScrapeConfig{
			Timeout:      10 * time.Second,
			MaxBodyBytes: 5242880,
		},
	}
}
