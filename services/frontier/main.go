// Package main is the entrypoint for the frontier BFS crawling and rate limiting service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/umutto/birdwatcher/services/frontier/internal/config"
	"github.com/umutto/birdwatcher/services/frontier/internal/consumer"
	"github.com/umutto/birdwatcher/services/frontier/internal/engine"
	"github.com/umutto/birdwatcher/services/frontier/internal/engine/valkey"
	"github.com/umutto/birdwatcher/services/frontier/internal/publisher"
	"github.com/umutto/birdwatcher/services/frontier/internal/server"
	"github.com/umutto/birdwatcher/shared/env"
	"github.com/umutto/birdwatcher/shared/logger"
	"github.com/umutto/birdwatcher/shared/nats"
)

// main sets up the signal trap, runs the service, and routes the process exit code.
func main() {
	ctx := context.Background()

	log := logger.New(logger.Config{
		Env:   env.Current(),
		Level: slog.LevelInfo,
	})

	if err := run(ctx, log); err != nil {
		log.ErrorContext(ctx, "frontier service terminated with error", slog.Any("error", err))
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
		"starting frontier service",
		slog.String("env", string(cfg.Env)),
		slog.Int("port", int(cfg.Server.Port)),
	)

	vStore, err := valkey.New(ctx, cfg.Valkey)
	if err != nil {
		return fmt.Errorf("initialize valkey: %w", err)
	}
	defer func() {
		if err := vStore.Close(); err != nil {
			log.ErrorContext(ctx, "close valkey store", slog.Any("error", err))
		}
	}()

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

	pub := publisher.New(nc)
	eng := engine.New(log, cfg.Crawl, vStore, pub)
	cons := consumer.New(log, eng, nc.JS)
	srv := server.New(log, cfg.Server, eng)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return srv.Start(gCtx)
	})
	g.Go(func() error {
		return cons.Start(gCtx)
	})
	g.Go(func() error {
		if err := eng.Run(gCtx); err != nil && !errors.Is(err, context.Canceled) {
			return fmt.Errorf("engine run: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run services: %w", err)
	}

	return nil
}
