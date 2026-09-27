package nats

import (
	"time"

	"github.com/nats-io/nats.go"
)

// Config defines connection and resilience settings for NATS.
type Config struct {
	URL           string
	Name          string
	MaxReconnects int
	ReconnectWait time.Duration
	Timeout       time.Duration
}

// DefaultConfig returns recommended settings for local development.
func DefaultConfig() Config {
	return Config{
		URL:           nats.DefaultURL,
		Name:          "",
		MaxReconnects: -1,
		ReconnectWait: 2 * time.Second,
		Timeout:       5 * time.Second,
	}
}
