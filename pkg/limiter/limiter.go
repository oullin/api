package limiter

import (
	"sync"
	"time"
)

// MemoryLimiter provides a simple in-memory failure-based rate limiter.
// It tracks failure timestamps per arbitrary key (e.g., ip|account)
// and decides whether the number of failures within a sliding window
// exceeds a configured threshold.
type MemoryLimiter struct {
	mu       sync.Mutex
	history  map[string][]time.Time // key -> failure timestamps
	window   time.Duration
	maxFails int
}

// NewMemoryLimiter constructs a MemoryLimiter with the specified sliding window duration
// and the maximum number of failures allowed within that window.
func NewMemoryLimiter(window time.Duration, maxFails int) *MemoryLimiter {
	return &MemoryLimiter{
		history:  make(map[string][]time.Time),
		window:   window,
		maxFails: maxFails,
	}
}

// TooMany reports whether the given key has reached or exceeded the
// maximum number of failures within the configured window.
func (r *MemoryLimiter) TooMany(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	slice := r.history[key]

	// prune old entries outside the window
	pruned := slice[:0]
	for _, t := range slice {
		if now.Sub(t) <= r.window {
			pruned = append(pruned, t)
		}
	}

	r.history[key] = pruned

	return len(pruned) >= r.maxFails
}

// Fail records a failure occurrence for the given key.
func (r *MemoryLimiter) Fail(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	r.history[key] = append(r.history[key], now)
}
