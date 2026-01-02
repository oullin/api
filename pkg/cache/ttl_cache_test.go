package cache_test

import (
	"testing"
	"time"

	"github.com/oullin/pkg/cache"
)

func TestTTLCache_UsedAndMark(t *testing.T) {
	c := cache.NewTTLCache()
	key := "acct|nonce1"
	if c.Used(key) {
		t.Fatalf("expected key not to be used initially")
	}
	c.Mark(key, 50*time.Millisecond)
	if !c.Used(key) {
		t.Fatalf("expected key to be marked as used within TTL")
	}
	time.Sleep(60 * time.Millisecond)
	if c.Used(key) {
		t.Fatalf("expected key to expire after TTL")
	}
}
