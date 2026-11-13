// Package main is the entrypoint for the stateless HTML fetcher and parser worker.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/umutto/birdwatcher/services/scraper/internal/config"
	"github.com/umutto/birdwatcher/services/scraper/internal/consumer"
	"github.com/umutto/birdwatcher/services/scraper/internal/fetcher"
	"github.com/umutto/birdwatcher/services/scraper/internal/publisher"
	"github.com/umutto/birdwatcher/shared/env"
	"github.com/umutto/birdwatcher/shared/logger"
	"github.com/umutto/birdwatcher/shared/nats"
)

// main initializes the logger, runs the service, and exits with a non-zero code on error.
func main() {
	ctx := context.Background()

	log := logger.New(logger.Config{
		Env:   env.Current(),
		Level: slog.LevelInfo,
	})

	if err := run(ctx, log); err != nil {
		log.ErrorContext(ctx, "scraper service terminated with error", slog.Any("error", err))
		os.Exit(1)
	}
}

// run coordinates configuration loading, dependency initialization, and graceful shutdown.
func run(ctx context.Context, log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log.InfoContext(
		ctx,
		"starting scraper service",
		slog.String("env", string(cfg.Env)),
	)

	nc, err := nats.Connect(ctx, nats.Config{URL: cfg.NATS.URL}, log)
	if err != nil {
		return fmt.Errorf("initialize nats: %w", err)
	}
	defer func() {
		if err := nc.Close(); err != nil {
			log.ErrorContext(ctx, "close nats connection", slog.Any("error", err))
		}
	}()

	if _, err := nc.CreateOrUpdateStream(ctx, nats.DefaultStreamConfig()); err != nil {
		return fmt.Errorf("create or update stream: %w", err)
	}

	fetch := fetcher.New(cfg.Scrape)
	pub := publisher.New(nc)
	cons := consumer.New(log, fetch, pub, nc.JS)

	if err := cons.Start(ctx); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}

	return nil
}
