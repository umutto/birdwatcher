// Package main is the entrypoint for the catalog API gateway and database service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/database"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/services/catalog/internal/server"
	"github.com/umutto/birdwatcher/shared/env"
	"github.com/umutto/birdwatcher/shared/logger"
)

// main sets up the signal trap, runs the service, and routes the process exit code.
func main() {
	ctx := context.Background()

	log := logger.New(logger.Config{
		Env:   env.Current(),
		Level: slog.LevelInfo,
	})

	if err := run(ctx, log); err != nil {
		log.ErrorContext(ctx, "catalog service terminated with error", slog.Any("error", err))
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
		ctx, "starting catalog service",
		slog.String("env", string(cfg.Env)),
		slog.Int("port", int(cfg.Server.Port)),
	)

	pool, err := database.New(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer func() {
		log.InfoContext(ctx, "closing database connection pool")
		pool.Close()
	}()

	repo := repository.New(pool)
	srv := server.New(log, repo, cfg.Server)

	if err := srv.Start(ctx); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}
