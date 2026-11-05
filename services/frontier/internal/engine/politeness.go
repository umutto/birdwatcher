package engine

import (
	"sync"
	"time"
)

// hostState tracks active concurrency and last access timestamp for a host.
type hostState struct {
	lastAccess time.Time
	inflight   int
}

// PolitenessLimiter enforces per-host request rate limits and concurrency caps.
type PolitenessLimiter struct {
	mu          sync.Mutex
	delay       time.Duration
	maxInflight int
	hosts       map[string]*hostState
}

// NewPolitenessLimiter creates a PolitenessLimiter with the specified config.
func NewPolitenessLimiter(delay time.Duration, maxInflight int) *PolitenessLimiter {
	return &PolitenessLimiter{
		delay:       delay,
		maxInflight: maxInflight,
		hosts:       make(map[string]*hostState),
	}
}

// Acquire reserves an execution slot for the host or returns the required cooldown duration.
func (p *PolitenessLimiter) Acquire(host string) (bool, time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	h := p.hosts[host]
	if h == nil {
		h = &hostState{}
		p.hosts[host] = h
	}

	if h.inflight >= p.maxInflight {
		return false, 0
	}

	elapsed := time.Since(h.lastAccess)
	if elapsed < p.delay {
		remaining := p.delay - elapsed
		return false, remaining
	}

	h.inflight++
	h.lastAccess = time.Now()

	return true, 0
}

// Release decrements the active inflight count for the specified host.
func (p *PolitenessLimiter) Release(host string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	h := p.hosts[host]
	if h != nil && h.inflight > 0 {
		h.inflight--
	}
}
