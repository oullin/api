package cache_test

import (
	"testing"
	"time"

	"github.com/oullin/pkg/cache"
)

// TestTTLCache_UseOnce verifies the behavior of UseOnce for first use,
// repeated use before expiry and reuse after the TTL has elapsed.
func TestTTLCache_UseOnce(t *testing.T) {
	t.Parallel()
	c := cache.NewTTLCache()
	key := "nonce"
	ttl := 100 * time.Millisecond

	t.Run("first use", func(t *testing.T) {
		if used := c.UseOnce(key, ttl); used {
			t.Fatalf("expected first UseOnce to return false")
		}
	})

	t.Run("second use before expiry", func(t *testing.T) {
		if used := c.UseOnce(key, ttl); !used {
			t.Fatalf("expected second UseOnce to return true before expiry")
		}
	})

	t.Run("use after expiry", func(t *testing.T) {
		time.Sleep(ttl + 50*time.Millisecond)
		if used := c.UseOnce(key, ttl); used {
			t.Fatalf("expected UseOnce to return false for an expired key")
		}
	})
}

// TestTTLCache_Mark_PrunesExpiredEntries ensures that calling Mark prunes
// any expired keys in the cache.
func TestTTLCache_Mark_PrunesExpiredEntries(t *testing.T) {
	c := cache.NewTTLCache()
	c.Mark("old", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	c.Mark("new", 10*time.Millisecond) // should prune "old"

	// Note: Since TTLCache fields are private, we can only verify behavior
	// by checking that expired entries don't show up as "Used"
	if c.Used("old") {
		t.Fatalf("expected expired key to be pruned from cache")
	}
}
