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

	if used := c.UseOnce(key, 50*time.Millisecond); used {
		t.Fatalf("expected first UseOnce to return false")
	}
	if used := c.UseOnce(key, 50*time.Millisecond); !used {
		t.Fatalf("expected second UseOnce to return true before expiry")
	}
	time.Sleep(60 * time.Millisecond)
	if used := c.UseOnce(key, 50*time.Millisecond); used {
		t.Fatalf("expected expired key to be usable again")
	}
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
