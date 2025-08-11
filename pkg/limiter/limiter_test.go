package limiter

import (
	"testing"
	"time"
)

func TestMemoryLimiter_BasicFlow(t *testing.T) {
	lim := NewMemoryLimiter(50*time.Millisecond, 3)
	key := "ip|acct"

	if lim.TooMany(key) {
		t.Fatalf("should not be limited initially")
	}

	lim.Fail(key)
	lim.Fail(key)
	if lim.TooMany(key) {
		t.Fatalf("should not be limited before reaching threshold")
	}

	lim.Fail(key)
	if !lim.TooMany(key) {
		t.Fatalf("should be limited after reaching threshold")
	}

	// Wait for window to slide and prune
	time.Sleep(60 * time.Millisecond)
	if lim.TooMany(key) {
		t.Fatalf("should not be limited after window passes")
	}
}
