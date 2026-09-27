// Package nats provides connection management and JetStream helpers for Birdwatcher microservices.
package nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Client wraps a core NATS connection and its modern JetStream context.
type Client struct {
	Conn *nats.Conn
	JS   jetstream.JetStream
	log  *slog.Logger
}

// Connect establishes a connection to the NATS cluster, initializes JetStream,
// and registers lifecycle logging handlers.
func Connect(ctx context.Context, cfg Config, log *slog.Logger) (*Client, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("nats: context canceled: %w", err)
	}

	if cfg.URL == "" {
		cfg.URL = nats.DefaultURL
	}

	opts := []nats.Option{
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.Timeout(cfg.Timeout),
	}

	if cfg.Name != "" {
		opts = append(opts, nats.Name(cfg.Name))
	}

	opts = append(opts, loggingOptions(ctx, log)...)

	nc, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats: connect to %q: %w", cfg.URL, err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("nats: init jetstream: %w", err)
	}

	return &Client{Conn: nc, JS: js, log: log}, nil
}

// Close drains in-flight messages and terminates the NATS connection.
func (c *Client) Close() error {
	if c.Conn != nil {
		return c.Conn.Drain()
	}
	return nil
}

// loggingOptions configures connection lifecycle event handlers with structured logging.
func loggingOptions(ctx context.Context, log *slog.Logger) []nats.Option {
	if log == nil {
		return nil
	}

	// Detach from the caller's cancellation/timeout while preserving trace context
	asyncCtx := context.WithoutCancel(ctx)

	return []nats.Option{
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.WarnContext(asyncCtx, "nats: disconnected", slog.Any("error", err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.InfoContext(asyncCtx, "nats: reconnected", slog.String("url", nc.ConnectedUrl()))
		}),
	}
}
