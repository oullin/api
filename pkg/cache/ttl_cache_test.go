package cache

import (
	"sync"
	"testing"
	"time"
)

func TestTTLCache_BasicMarkAndUsed(t *testing.T) {
	c := NewTTLCache()

	// Unknown key should be false
	if c.Used("missing") {
		t.Fatalf("expected missing key to be false")
	}

	// Mark a key with a short TTL and verify it's true before expiry
	c.Mark("k1", 50*time.Millisecond)
	if !c.Used("k1") {
		t.Fatalf("expected k1 to be true before expiry")
	}

	// Sleep past TTL and verify it expires
	time.Sleep(60 * time.Millisecond)
	if c.Used("k1") {
		t.Fatalf("expected k1 to be false after expiry")
	}

	// After lazy pruning, subsequent calls should also be false
	if c.Used("k1") {
		t.Fatalf("expected k1 to remain false after pruning")
	}
}

func TestTTLCache_IndependentKeys(t *testing.T) {
	c := NewTTLCache()

	c.Mark("short", 20*time.Millisecond)
	c.Mark("long", 200*time.Millisecond)

	// Immediately both should be usable
	if !c.Used("short") || !c.Used("long") {
		t.Fatalf("expected both keys to be true initially")
	}

	// Wait for the short to expire but not the long
	time.Sleep(50 * time.Millisecond)

	if c.Used("short") {
		t.Fatalf("expected short to be expired")
	}
	if !c.Used("long") {
		t.Fatalf("expected long to still be valid")
	}
}

func TestTTLCache_RefreshTTL(t *testing.T) {
	c := NewTTLCache()

	c.Mark("k", 20*time.Millisecond)
	// Refresh before expiry
	time.Sleep(10 * time.Millisecond)
	c.Mark("k", 40*time.Millisecond)

	// Wait past the first TTL but before the refreshed TTL
	time.Sleep(15 * time.Millisecond) // total 25ms from first mark

	if !c.Used("k") {
		t.Fatalf("expected k to be valid due to refreshed TTL")
	}

	// Now wait past the refreshed TTL
	time.Sleep(30 * time.Millisecond) // total ~55ms from first mark
	if c.Used("k") {
		t.Fatalf("expected k to be expired after refreshed TTL")
	}
}

func TestTTLCache_Concurrency(t *testing.T) {
	c := NewTTLCache()

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			key := "k" // all share same key to contend on lock
			c.Mark(key, 200*time.Millisecond)
			if !c.Used(key) {
				t.Errorf("expected key to be usable; goroutine %d", i)
			}
		}(i)
	}

	wg.Wait()

	// After all goroutines, key should still be valid
	if !c.Used("k") {
		t.Fatalf("expected key to still be valid after concurrent access")
	}
}
