package engine_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/frontier/internal/config"
	"github.com/umutto/birdwatcher/services/frontier/internal/engine"
)

// fetchRecord captures arguments passed to PublishFetch.
type fetchRecord struct {
	jobID string
	url   string
	depth int32
}

// mockPublisher records publishing events in memory for test verification.
type mockPublisher struct {
	mu        sync.Mutex
	fetches   []fetchRecord
	completed []string
}

func newMockPublisher() *mockPublisher {
	return &mockPublisher{}
}

func (m *mockPublisher) PublishFetch(_ context.Context, jobID, targetURL string, depth int32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fetches = append(m.fetches, fetchRecord{jobID: jobID, url: targetURL, depth: depth})
	return nil
}

func (m *mockPublisher) PublishJobCompleted(_ context.Context, jobID, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.completed = append(m.completed, jobID)
	return nil
}

// mockVisitedStore tracks visited URLs in memory for test verification.
type mockVisitedStore struct {
	mu      sync.Mutex
	visited map[string]map[string]struct{}
}

func newMockVisitedStore() *mockVisitedStore {
	return &mockVisitedStore{
		visited: make(map[string]map[string]struct{}),
	}
}

func (m *mockVisitedStore) Add(_ context.Context, jobID, normalizedURL string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	urls, ok := m.visited[jobID]
	if !ok {
		urls = make(map[string]struct{})
		m.visited[jobID] = urls
	}

	if _, exists := urls[normalizedURL]; exists {
		return false, nil
	}
	urls[normalizedURL] = struct{}{}
	return true, nil
}

func (m *mockVisitedStore) Has(_ context.Context, jobID, normalizedURL string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	urls := m.visited[jobID]
	if urls == nil {
		return false, nil
	}
	_, ok := urls[normalizedURL]
	return ok, nil
}

func (m *mockVisitedStore) Count(_ context.Context, jobID string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.visited[jobID])), nil
}

func (m *mockVisitedStore) RemoveJob(_ context.Context, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.visited, jobID)
	return nil
}

func newTestEngine() (*engine.Engine, *mockVisitedStore, *mockPublisher) {
	discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := config.CrawlConfig{
		HostDelay:          10 * time.Millisecond,
		MaxInflightPerHost: 2,
	}
	visited := newMockVisitedStore()
	publisher := newMockPublisher()
	eng := engine.New(discardLogger, cfg, visited, publisher)
	return eng, visited, publisher
}

func TestEngine_StartJob(t *testing.T) {
	t.Parallel()

	t.Run("enqueues seed url and marks visited", func(t *testing.T) {
		t.Parallel()
		eng, visited, _ := newTestEngine()
		ctx := context.Background()

		if err := eng.StartJob(ctx, "job-1", "https://example.com", 2, 10); err != nil {
			t.Fatalf("StartJob() error = %v, want nil", err)
		}

		if got := eng.QueueLen(); got != 1 {
			t.Errorf("QueueLen() = %d, want 1", got)
		}

		has, err := visited.Has(ctx, "job-1", "https://example.com/")
		if err != nil || !has {
			t.Errorf("visited.Has() = %v, err = %v, want true, nil", has, err)
		}
	})

	t.Run("rejects duplicate job id", func(t *testing.T) {
		t.Parallel()
		eng, _, _ := newTestEngine()
		ctx := context.Background()

		if err := eng.StartJob(ctx, "job-1", "https://example.com", 2, 10); err != nil {
			t.Fatalf("initial StartJob() error = %v, want nil", err)
		}

		if err := eng.StartJob(ctx, "job-1", "https://example.com", 2, 10); err == nil {
			t.Errorf("duplicate StartJob() error = nil, want error")
		}
	})
}

func TestEngine_DispatchNext(t *testing.T) {
	t.Parallel()

	t.Run("dispatches queued url to publisher", func(t *testing.T) {
		t.Parallel()
		eng, _, publisher := newTestEngine()
		ctx := context.Background()

		if err := eng.StartJob(ctx, "job-1", "https://example.com", 2, 10); err != nil {
			t.Fatalf("StartJob() error = %v, want nil", err)
		}

		dispatched, err := eng.DispatchNext(ctx)
		if err != nil {
			t.Fatalf("DispatchNext() error = %v, want nil", err)
		}
		if !dispatched {
			t.Errorf("DispatchNext() = false, want true")
		}

		if got := eng.QueueLen(); got != 0 {
			t.Errorf("QueueLen() = %d, want 0", got)
		}

		publisher.mu.Lock()
		defer publisher.mu.Unlock()
		if len(publisher.fetches) != 1 {
			t.Fatalf("len(publisher.fetches) = %d, want 1", len(publisher.fetches))
		}
		if got := publisher.fetches[0].url; got != "https://example.com/" {
			t.Errorf("fetches[0].url = %q, want %q", got, "https://example.com/")
		}
	})

	t.Run("returns false on empty queue", func(t *testing.T) {
		t.Parallel()
		eng, _, _ := newTestEngine()
		ctx := context.Background()

		dispatched, err := eng.DispatchNext(ctx)
		if err != nil {
			t.Fatalf("DispatchNext() error = %v, want nil", err)
		}
		if dispatched {
			t.Errorf("DispatchNext() = true, want false")
		}
	})
}

func TestEngine_HandleDiscoveredURLs(t *testing.T) {
	t.Parallel()

	t.Run("filters external hosts", func(t *testing.T) {
		t.Parallel()
		eng, _, _ := newTestEngine()
		ctx := context.Background()

		if err := eng.StartJob(ctx, "job-1", "https://example.com", 2, 10); err != nil {
			t.Fatalf("StartJob() error = %v, want nil", err)
		}
		if _, err := eng.DispatchNext(ctx); err != nil {
			t.Fatalf("DispatchNext() error = %v, want nil", err)
		}

		discovered := []string{
			"https://example.com/about",
			"https://otherdomain.com/page",
		}
		if err := eng.HandleDiscoveredURLs(ctx, "job-1", "https://example.com", 0, discovered); err != nil {
			t.Fatalf("HandleDiscoveredURLs() error = %v, want nil", err)
		}

		if got := eng.QueueLen(); got != 1 {
			t.Errorf("QueueLen() = %d, want 1", got)
		}
	})

	t.Run("completes job when queue and inflight empty", func(t *testing.T) {
		t.Parallel()
		eng, _, publisher := newTestEngine()
		ctx := context.Background()

		if err := eng.StartJob(ctx, "job-1", "https://example.com", 2, 10); err != nil {
			t.Fatalf("StartJob() error = %v, want nil", err)
		}
		if _, err := eng.DispatchNext(ctx); err != nil {
			t.Fatalf("DispatchNext() error = %v, want nil", err)
		}

		if err := eng.HandleDiscoveredURLs(ctx, "job-1", "https://example.com", 0, nil); err != nil {
			t.Fatalf("HandleDiscoveredURLs() error = %v, want nil", err)
		}

		publisher.mu.Lock()
		defer publisher.mu.Unlock()
		if len(publisher.completed) != 1 || publisher.completed[0] != "job-1" {
			t.Errorf("publisher.completed = %v, want [job-1]", publisher.completed)
		}
	})
}

func TestEngine_Run(t *testing.T) {
	t.Parallel()

	t.Run("stops on context cancellation", func(t *testing.T) {
		t.Parallel()
		eng, _, _ := newTestEngine()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := eng.Run(ctx)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run() error = %v, want %v", err, context.Canceled)
		}
	})
}
