package config_test

import (
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/frontier/internal/config"
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
			name: "empty nats url",
			mutate: func(c *config.Config) {
				c.NATS.URL = " "
			},
			wantErr: true,
		},
		{
			name: "negative host delay",
			mutate: func(c *config.Config) {
				c.Crawl.HostDelay = -1
			},
			wantErr: true,
		},
		{
			name: "negative max inflight per host",
			mutate: func(c *config.Config) {
				c.Crawl.MaxInflightPerHost = -1
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

func validConfig() *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Port: 9090,
		},
		NATS: config.NATSConfig{
			URL: "nats://localhost:4222",
		},
		Crawl: config.CrawlConfig{
			HostDelay:          250 * time.Millisecond,
			MaxInflightPerHost: 2,
		},
	}
}
