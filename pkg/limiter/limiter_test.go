package limiter

import (
	"testing"
	"time"
)

// TestTooManyThreshold verifies that TooMany returns true only after
// maxFails failures have been recorded within the window.
func TestTooManyThreshold(t *testing.T) {
	r := NewMemoryLimiter(1*time.Minute, 3)
	key := "ip|account"

	if r.TooMany(key) {
		t.Fatalf("expected TooMany to be false before any failures")
	}

	r.Fail(key)
	if r.TooMany(key) {
		t.Fatalf("expected TooMany to be false after 1 failure (< maxFails)")
	}

	r.Fail(key)
	if r.TooMany(key) {
		t.Fatalf("expected TooMany to be false after 2 failures (< maxFails)")
	}

	r.Fail(key)
	if !r.TooMany(key) {
		t.Fatalf("expected TooMany to be true after reaching maxFails")
	}
}

// TestWindowPruning verifies that failures older than the window are pruned
// and do not contribute to the TooMany decision.
func TestWindowPruning(t *testing.T) {
	// Use a short window to make the test fast and deterministic
	window := 50 * time.Millisecond
	r := NewMemoryLimiter(window, 2)
	key := "client|user"

	// Record one failure and wait for it to expire
	r.Fail(key)
	time.Sleep(window + 10*time.Millisecond)

	// Calling TooMany triggers pruning of the older entry
	if r.TooMany(key) {
		t.Fatalf("expected TooMany to be false after window expiration and pruning")
	}

	// Now add two quick failures within the window
	r.Fail(key)
	r.Fail(key)

	if !r.TooMany(key) {
		t.Fatalf("expected TooMany to be true after 2 failures within window (maxFails=2)")
	}
}

// TestIndependentKeys checks that limiter maintains separate counters per key.
func TestIndependentKeys(t *testing.T) {
	r := NewMemoryLimiter(1*time.Second, 2)
	keyA := "ipA|acct"
	keyB := "ipB|acct"

	r.Fail(keyA)
	if r.TooMany(keyA) {
		t.Fatalf("TooMany should be false for keyA with 1 failure")
	}

	// keyB should be unaffected by keyA's failures
	if r.TooMany(keyB) { // triggers prune/check but no failures recorded for keyB
		t.Fatalf("TooMany should be false for keyB with 0 failures")
	}

	// Push keyB over the threshold independently
	r.Fail(keyB)
	r.Fail(keyB)

	if !r.TooMany(keyB) {
		t.Fatalf("TooMany should be true for keyB after reaching maxFails")
	}

	// keyA still below threshold
	if r.TooMany(keyA) {
		t.Fatalf("TooMany should still be false for keyA (only 1 failure)")
	}
}
