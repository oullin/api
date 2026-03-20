package cache

import (
	"sync"
	"time"
)

// TTLCache is a tiny in-memory TTL key store.
// Values are not stored; only key existence within TTL is tracked.
// It is process-local and NOT suitable for distributed deployments.
// Use a shared cache (e.g., Redis) for multi-instance setups.
type TTLCache struct {
	mu   sync.Mutex
	data map[string]time.Time // key -> expiry time
}

// NewTTLCache creates a new empty TTL cache.
func NewTTLCache() *TTLCache {
	return &TTLCache{
		data: make(map[string]time.Time),
	}
}

// Used reports whether key is present and not expired.
// It lazily prunes expired entries on access.
func (c *TTLCache) Used(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	exp, ok := c.data[key]
	if !ok {
		return false
	}

	if time.Now().After(exp) {
		delete(c.data, key)
		return false
	}

	return true
}

// Mark stores the key with a time-to-live.
func (c *TTLCache) Mark(key string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	// Opportunistic prune of expired entries to bound memory growth

	for k, exp := range c.data {
		if now.After(exp) {
			delete(c.data, k)
		}
	}

	c.data[key] = now.Add(ttl)
}

// UseOnce atomically checks whether the key has already been used and, if not,
// marks it as used with the provided ttl. Returns true if the key was already
// present (and not expired), false if it was newly marked.
func (c *TTLCache) UseOnce(key string, ttl time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if exp, ok := c.data[key]; ok && now.Before(exp) {
		return true // already used
	}

	c.data[key] = now.Add(ttl)

	return false
}
