// Package valkey provides a Valkey-backed store for crawler state and URL deduplication.
package valkey

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/umutto/birdwatcher/services/frontier/internal/config"
)

const prefix = "visited:"

// Store provides visited URL tracking backed by Valkey.
type Store struct {
	client *redis.Client
}

// New connects to Valkey and verifies connectivity.
func New(ctx context.Context, cfg config.ValkeyConfig) (*Store, error) {
	client := redis.NewClient(&redis.Options{Addr: cfg.Addr})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping valkey: %w", err)
	}

	return &Store{client: client}, nil
}

// Add atomically marks a normalized URL as visited for the specified job.
// It returns true if the URL was newly recorded, or false if it was already visited.
func (s *Store) Add(ctx context.Context, jobID, normalizedURL string) (bool, error) {
	key := prefix + jobID

	n, err := s.client.SAdd(ctx, key, normalizedURL).Result()
	if err != nil {
		return false, fmt.Errorf("valkey sadd %s: %w", key, err)
	}

	return n > 0, nil
}

// Has checks whether a normalized URL has already been visited for the specified job.
func (s *Store) Has(ctx context.Context, jobID, normalizedURL string) (bool, error) {
	key := prefix + jobID

	m, err := s.client.SIsMember(ctx, key, normalizedURL).Result()
	if err != nil {
		return false, fmt.Errorf("valkey sismember %s: %w", key, err)
	}

	return m, nil
}

// Count returns the total number of unique visited URLs recorded for the specified job.
func (s *Store) Count(ctx context.Context, jobID string) (int64, error) {
	key := prefix + jobID

	c, err := s.client.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("valkey scard %s: %w", key, err)
	}

	return c, nil
}

// RemoveJob deletes all visited URL records associated with the specified job.
func (s *Store) RemoveJob(ctx context.Context, jobID string) error {
	key := prefix + jobID

	err := s.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("valkey del %s: %w", key, err)
	}

	return nil
}

// Close terminates the Valkey client connection pool.
func (s *Store) Close() error {
	return s.client.Close()
}
