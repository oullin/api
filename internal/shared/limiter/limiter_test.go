package limiter_test

import (
	"testing"
	"time"

	"github.com/oullin/internal/shared/limiter"
)

func TestMemoryLimiter_BasicFlow(t *testing.T) {
	lim := limiter.NewMemoryLimiter(50*time.Millisecond, 3)
	key := "ip|acct"

	if lim.TooMany(key) {
		t.Fatalf("expected key not to be limited initially")
	}

	lim.Fail(key)
	lim.Fail(key)
	if lim.TooMany(key) {
		t.Fatalf("expected key not to be limited before reaching threshold")
	}

	lim.Fail(key)
	if !lim.TooMany(key) {
		t.Fatalf("expected key to be limited after reaching threshold")
	}

	// Wait for window to slide and prune
	time.Sleep(60 * time.Millisecond)
	if lim.TooMany(key) {
		t.Fatalf("expected key not to be limited after window passes")
	}
}
