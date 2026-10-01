package nats

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	// DefaultStreamName is the primary JetStream stream name.
	DefaultStreamName = "BIRDWATCHER"
	// DefaultStreamSubject matches all v1 Birdwatcher pipeline events.
	DefaultStreamSubject = "birdwatcher.v1.>"
)

// DefaultStreamConfig returns the recommended JetStream stream configuration for the Birdwatcher pipeline.
func DefaultStreamConfig() jetstream.StreamConfig {
	return jetstream.StreamConfig{
		Name:       DefaultStreamName,
		Subjects:   []string{DefaultStreamSubject},
		Storage:    jetstream.FileStorage,
		Retention:  jetstream.LimitsPolicy,
		Discard:    jetstream.DiscardOld,
		Duplicates: 2 * time.Minute,
	}
}

// CreateOrUpdateStream idempotently provisions or updates a JetStream stream.
func (c *Client) CreateOrUpdateStream(ctx context.Context, cfg jetstream.StreamConfig) (jetstream.Stream, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("nats: context cancelled: %w", err)
	}

	if c == nil || c.JS == nil {
		return nil, errors.New("nats: jetstream not initialized")
	}

	stream, err := c.JS.CreateOrUpdateStream(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("nats: create or update stream %q: %w", cfg.Name, err)
	}

	c.log.InfoContext(ctx, "nats: stream created or updated", slog.String("stream", cfg.Name))

	return stream, nil
}
