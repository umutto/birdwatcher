// Package engine implements the BFS URL queue, deduplication and politeness engine.
package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/umutto/birdwatcher/services/frontier/internal/config"
)

// defaultDispatchInterval is the polling frequency of the engine's background dispatch loop.
const defaultDispatchInterval = 25 * time.Millisecond

// VisitedStore manages visited URL deduplication across crawl jobs.
type VisitedStore interface {
	Add(ctx context.Context, jobID, normalizedURL string) (bool, error)
	Has(ctx context.Context, jobID, normalizedURL string) (bool, error)
	Count(ctx context.Context, jobID string) (int64, error)
	RemoveJob(ctx context.Context, jobID string) error
}

// Publisher publishes crawl task events to the messaging transport.
type Publisher interface {
	PublishFetch(ctx context.Context, jobID, targetURL string, depth int32) error
	PublishJobCompleted(ctx context.Context, jobID string, errMsg string) error
}

// jobState tracks traversal bounds, progress counters, and active tasks for a crawl job.
type jobState struct {
	host       string
	maxDepth   int32
	maxPages   int32
	dispatched int32
	inflight   int32
}

// Engine coordinates crawl jobs, URL deduplication, and politeness scheduling.
type Engine struct {
	log       *slog.Logger
	cfg       config.CrawlConfig
	visited   VisitedStore
	publisher Publisher

	mu      sync.Mutex
	queue   *Queue
	limiter *PolitenessLimiter
	jobs    map[string]*jobState
}

// New creates an initialized crawl Engine.
func New(log *slog.Logger, cfg config.CrawlConfig, visited VisitedStore, publisher Publisher) *Engine {
	return &Engine{
		log:       log,
		cfg:       cfg,
		visited:   visited,
		publisher: publisher,
		queue:     NewQueue(),
		limiter:   NewPolitenessLimiter(cfg.HostDelay, int(cfg.MaxInflightPerHost)),
		jobs:      make(map[string]*jobState),
	}
}

// StartJob validates the root URL, registers job state, records the root as visited, and enqueues it.
func (e *Engine) StartJob(ctx context.Context, jobID, rootURL string, maxDepth, maxPages int32) error {
	normalizedURL, err := NormalizeURL(rootURL)
	if err != nil {
		return fmt.Errorf("normalize root url %q: %w", rootURL, err)
	}

	parsed, err := url.Parse(normalizedURL)
	if err != nil {
		return fmt.Errorf("parse root url %q: %w", normalizedURL, err)
	}

	e.mu.Lock()
	if _, ok := e.jobs[jobID]; ok {
		e.mu.Unlock()
		return fmt.Errorf("job %q already exists", jobID)
	}
	e.jobs[jobID] = &jobState{
		host:     parsed.Host,
		maxDepth: maxDepth,
		maxPages: maxPages,
	}
	e.mu.Unlock()

	if _, err := e.visited.Add(ctx, jobID, normalizedURL); err != nil {
		e.mu.Lock()
		delete(e.jobs, jobID)
		e.mu.Unlock()
		return fmt.Errorf("record visited root url: %w", err)
	}
	e.queue.Push(Item{
		JobID: jobID,
		URL:   normalizedURL,
		Depth: 0,
	})

	return nil
}

// HandleDiscoveredURLs filters discovered links, deduplicates them against VisitedStore,
// pushes new URLs to the queue, and checks for job completion.
func (e *Engine) HandleDiscoveredURLs(
	ctx context.Context,
	jobID,
	parentURL string,
	parentDepth int32,
	discoveredURLs []string,
) error {
	e.mu.Lock()
	job, ok := e.jobs[jobID]
	if !ok {
		e.mu.Unlock()
		return fmt.Errorf("job %q not found", jobID)
	}

	if job.inflight > 0 {
		job.inflight--
	}

	atLimit := job.dispatched >= job.maxPages || parentDepth+1 > job.maxDepth
	e.mu.Unlock()

	e.limiter.Release(job.host)

	if !atLimit {
		for _, raw := range discoveredURLs {
			valid, err := validateDiscoveredURL(raw, job.host)
			if err != nil {
				e.log.DebugContext(
					ctx,
					"skipping invalid url",
					slog.String("parent_url", parentURL),
					slog.String("url", raw),
					slog.Any("error", err),
				)
				continue
			}

			isNew, err := e.visited.Add(ctx, jobID, valid)
			if err != nil {
				return fmt.Errorf("record visited url: %w", err)
			}
			if !isNew {
				continue
			}

			e.queue.Push(Item{
				JobID: jobID,
				URL:   valid,
				Depth: parentDepth + 1,
			})
		}
	}

	e.mu.Lock()
	isDone := job.inflight == 0 && e.queue.Len() == 0
	if isDone {
		delete(e.jobs, jobID)
	}
	e.mu.Unlock()

	if isDone {
		if err := e.publisher.PublishJobCompleted(ctx, jobID, ""); err != nil {
			return fmt.Errorf("publish job completed: %w", err)
		}
		return nil
	}

	return nil
}

// Run starts the dispatch loop, periodically pumping queued URLs to the fetch publisher until ctx is cancelled.
func (e *Engine) Run(ctx context.Context) error {
	ticker := time.NewTicker(defaultDispatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if _, err := e.dispatchNext(ctx); err != nil {
				e.log.WarnContext(ctx, "dispatch failed", slog.Any("error", err))
			}
		}
	}
}

// validateDiscoveredURL normalizes a candidate link and verifies it matches the target job host.
func validateDiscoveredURL(rawURL, jobHost string) (string, error) {
	normalized, err := NormalizeURL(rawURL)
	if err != nil {
		return "", fmt.Errorf("normalize url: %w", err)
	}

	parsed, err := url.Parse(normalized)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}
	if parsed.Host != jobHost {
		return "", errors.New("external host")
	}

	return normalized, nil
}

// dispatchNext pops the next URL candidate, checks politeness and budgets, and publishes a fetch task.
func (e *Engine) dispatchNext(ctx context.Context) (bool, error) {
	item, ok := e.queue.Pop()
	if !ok {
		return false, nil
	}

	e.mu.Lock()
	job, ok := e.jobs[item.JobID]
	if !ok {
		e.mu.Unlock()
		return false, nil
	}
	if job.dispatched >= job.maxPages {
		isDone := job.inflight == 0 && e.queue.Len() == 0
		if isDone {
			delete(e.jobs, item.JobID)
		}
		e.mu.Unlock()

		if isDone {
			if err := e.publisher.PublishJobCompleted(ctx, item.JobID, ""); err != nil {
				return false, fmt.Errorf("publish job completed: %w", err)
			}
		}

		return false, nil
	}

	allowed, _ := e.limiter.Acquire(job.host)
	if !allowed {
		e.mu.Unlock()
		e.queue.Push(item)
		return false, nil
	}

	job.dispatched++
	job.inflight++
	e.mu.Unlock()

	if err := e.publisher.PublishFetch(ctx, item.JobID, item.URL, item.Depth); err != nil {
		e.mu.Lock()
		if j, ok := e.jobs[item.JobID]; ok {
			j.dispatched--
			j.inflight--
		}
		e.mu.Unlock()
		e.limiter.Release(job.host)
		return false, fmt.Errorf("publish fetch: %w", err)
	}

	return true, nil
}
