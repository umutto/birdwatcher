// Package engine implements the BFS URL queue, deduplication and politeness engine.
package engine

import "context"

// VisitedStore manages visited URL deduplication across crawl jobs.
type VisitedStore interface {
	Add(ctx context.Context, jobID, normalizedURL string) (bool, error)
	Has(ctx context.Context, jobID, normalizedURL string) (bool, error)
	Count(ctx context.Context, jobID string) (int64, error)
	RemoveJob(ctx context.Context, jobID string) error
}
