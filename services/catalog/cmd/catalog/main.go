// Package main is the entrypoint for the catalog API gateway and database service.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	birdwatcherv1 "github.com/umutto/birdwatcher/gen/go/birdwatcher/v1"
	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/consumer"
	"github.com/umutto/birdwatcher/services/catalog/internal/database"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/services/catalog/internal/server"
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
		ctx,
		"starting catalog service",
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

	cc, err := grpc.NewClient(
		cfg.Server.Frontier.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("initialize frontier grpc connection: %w", err)
	}
	defer func() {
		if err := cc.Close(); err != nil {
			log.ErrorContext(ctx, "close frontier grpc connection", slog.Any("error", err))
		}
	}()

	frontier := birdwatcherv1.NewFrontierServiceClient(cc)

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

	srv := server.New(log, cfg.Server, pool, repo, frontier)
	cons := consumer.New(log, repo, nc.JS)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return srv.Start(gCtx)
	})
	g.Go(func() error {
		return cons.Start(gCtx)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("run services: %w", err)
	}

	return nil
}
