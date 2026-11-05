package engine_test

import (
	"sync"
	"testing"
	"time"

	"github.com/umutto/birdwatcher/services/frontier/internal/engine"
)

func TestPoliteness(t *testing.T) {
	t.Run("allows initial acquire for new host", func(t *testing.T) {
		p := engine.NewPolitenessLimiter(100*time.Millisecond, 2)

		ok, wait := p.Acquire("example.com")
		if !ok || wait != 0 {
			t.Fatalf("Acquire() = (%v, %v); want (true, 0)", ok, wait)
		}
	})

	t.Run("enforces delay between requests", func(t *testing.T) {
		delay := 5 * time.Millisecond
		p := engine.NewPolitenessLimiter(delay, 2)

		ok, _ := p.Acquire("example.com")
		if !ok {
			t.Fatal("first Acquire() failed; want true")
		}

		ok, wait := p.Acquire("example.com")
		if ok {
			t.Fatal("second Acquire() succeeded immediately; want false")
		}
		if wait <= 0 || wait > delay {
			t.Fatalf("wait duration = %v; want > 0 and <= %v", wait, delay)
		}

		time.Sleep(wait)

		ok, _ = p.Acquire("example.com")
		if !ok {
			t.Fatal("Acquire() after delay cooldown failed; want true")
		}
	})

	t.Run("enforces max inflight limit and allows acquire after release", func(t *testing.T) {
		p := engine.NewPolitenessLimiter(0, 2)

		if ok, _ := p.Acquire("example.com"); !ok {
			t.Fatal("acquire 1 failed; want true")
		}
		if ok, _ := p.Acquire("example.com"); !ok {
			t.Fatal("acquire 2 failed; want true")
		}

		ok, wait := p.Acquire("example.com")
		if ok || wait != 0 {
			t.Fatalf("acquire 3 = (%v, %v); want (false, 0) due to inflight saturation", ok, wait)
		}

		p.Release("example.com")

		ok, _ = p.Acquire("example.com")
		if !ok {
			t.Fatal("acquire after release failed; want true")
		}
	})

	t.Run("isolates rate limits between different hosts", func(t *testing.T) {
		p := engine.NewPolitenessLimiter(100*time.Millisecond, 1)

		if ok, _ := p.Acquire("a.com"); !ok {
			t.Fatal("Acquire(a.com) failed; want true")
		}

		ok, wait := p.Acquire("b.com")
		if !ok || wait != 0 {
			t.Fatalf("Acquire(b.com) = (%v, %v); want (true, 0)", ok, wait)
		}
	})

	t.Run("release on unknown host is a safe no-op", func(t *testing.T) {
		p := engine.NewPolitenessLimiter(100*time.Millisecond, 2)
		p.Release("unknown.com")
	})

	t.Run("handles concurrent acquires and releases safely", func(t *testing.T) {
		p := engine.NewPolitenessLimiter(1*time.Millisecond, 2)
		const numWorkers = 50

		var wg sync.WaitGroup
		for range numWorkers {
			wg.Go(func() {
				ok, _ := p.Acquire("example.com")
				if ok {
					p.Release("example.com")
				}
			})
		}
		wg.Wait()
	})
}
