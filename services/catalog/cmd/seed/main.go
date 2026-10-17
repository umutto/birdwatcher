// Package main is the database seed CLI tool for populating bird taxonomy and sample sightings.
package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/umutto/birdwatcher/services/catalog/internal/config"
	"github.com/umutto/birdwatcher/services/catalog/internal/database"
	"github.com/umutto/birdwatcher/services/catalog/internal/repository"
	"github.com/umutto/birdwatcher/shared/env"
	"github.com/umutto/birdwatcher/shared/logger"
)

// speciesJSON holds the embedded Audubon bird taxonomy dataset.
//
//go:embed data/species.json
var speciesJSON []byte

// main initializes the logger, executes the seed workflow, and exits with a non-zero status on error.
func main() {
	ctx := context.Background()

	log := logger.New(logger.Config{
		Env:   env.Current(),
		Level: slog.LevelInfo,
	})

	if err := run(ctx, log); err != nil {
		log.ErrorContext(ctx, "catalog seeder terminated with error", slog.Any("error", err))
		os.Exit(1)
	}
}

// run coordinates configuration loading, database initialization, and seed execution.
func run(ctx context.Context, log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	pool, err := database.New(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer func() {
		log.InfoContext(ctx, "closing database connection pool")
		pool.Close()
	}()

	if err := seedSpecies(ctx, log, pool); err != nil {
		return fmt.Errorf("seed species: %w", err)
	}

	return nil
}

// seedSpecies parses the embedded taxonomy JSON and upserts all species records within a single transaction.
func seedSpecies(ctx context.Context, log *slog.Logger, pool *pgxpool.Pool) error {
	var species []repository.UpsertSpeciesParams
	if err := json.Unmarshal(speciesJSON, &species); err != nil {
		return fmt.Errorf("unmarshal species json: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	qtx := repository.New(tx)
	for _, s := range species {
		if s.Aliases == nil {
			s.Aliases = []string{}
		}
		if len(s.Metadata) == 0 {
			s.Metadata = json.RawMessage(`{}`)
		}
		if _, err := qtx.UpsertSpecies(ctx, s); err != nil {
			return fmt.Errorf("upsert species plate %d: %w", s.PlateNumber, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	log.InfoContext(ctx, "seeded species", slog.Int("count", len(species)))

	return nil
}
