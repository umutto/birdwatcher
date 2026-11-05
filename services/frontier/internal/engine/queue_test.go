package engine_test

import (
	"strconv"
	"sync"
	"testing"

	"github.com/umutto/birdwatcher/services/frontier/internal/engine"
)

func TestQueue(t *testing.T) {
	t.Run("preserves fifo ordering", func(t *testing.T) {
		q := engine.NewQueue()

		items := []engine.Item{
			{JobID: "job-1", URL: "https://example.com/1", Depth: 0},
			{JobID: "job-1", URL: "https://example.com/2", Depth: 1},
			{JobID: "job-1", URL: "https://example.com/3", Depth: 2},
		}

		for _, item := range items {
			q.Push(item)
		}

		if got := q.Len(); got != len(items) {
			t.Fatalf("q.Len() = %d; want %d", got, len(items))
		}

		for i, want := range items {
			got, ok := q.Pop()
			if !ok {
				t.Fatalf("pop [%d]: expected ok = true, got false", i)
			}
			if got != want {
				t.Errorf("pop [%d] = %+v; want %+v", i, got, want)
			}
		}

		if got := q.Len(); got != 0 {
			t.Errorf("q.Len() after draining = %d; want 0", got)
		}
	})

	t.Run("returns false and empty item when queue is empty", func(t *testing.T) {
		q := engine.NewQueue()

		item, ok := q.Pop()
		if ok {
			t.Fatalf("q.Pop() on empty queue returned ok = true; want false")
		}
		if item != (engine.Item{}) {
			t.Fatalf("q.Pop() on empty queue returned %+v; want empty Item{}", item)
		}
		if got := q.Len(); got != 0 {
			t.Errorf("q.Len() on empty queue = %d; want 0", got)
		}
	})

	t.Run("handles concurrent pushes and pops safely", func(t *testing.T) {
		q := engine.NewQueue()
		const numWorkers = 50

		var wg sync.WaitGroup
		for i := range numWorkers {
			wg.Go(func() {
				q.Push(engine.Item{
					JobID: "job-" + strconv.Itoa(i),
					URL:   "https://example.com/" + strconv.Itoa(i),
					Depth: 0,
				})
			})
		}
		for range numWorkers {
			wg.Go(func() { q.Pop() })
		}

		wg.Wait()
	})
}
