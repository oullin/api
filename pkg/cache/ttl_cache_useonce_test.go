package cache

import (
	"testing"
	"time"
)

// TestTTLCache_UseOnce verifies the behavior of UseOnce for first use,
// repeated use before expiry and reuse after the TTL has elapsed.
func TestTTLCache_UseOnce(t *testing.T) {
	c := NewTTLCache()
	key := "nonce"
	ttl := 50 * time.Millisecond

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
		time.Sleep(ttl + 10*time.Millisecond)
		if used := c.UseOnce(key, ttl); used {
			t.Fatalf("expected UseOnce to return false for an expired key")
		}
	})
}

// TestTTLCache_Mark_PrunesExpiredEntries ensures that calling Mark prunes
// any expired keys in the cache.
func TestTTLCache_Mark_PrunesExpiredEntries(t *testing.T) {
	c := NewTTLCache()
	c.Mark("old", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	c.Mark("new", 10*time.Millisecond) // should prune "old"

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.data["old"]; ok {
		t.Fatalf("expected expired key to be pruned from cache")
	}
}
