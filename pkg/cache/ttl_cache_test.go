package cache

import (
	"testing"
	"time"
)

func TestTTLCache_UsedAndMark(t *testing.T) {
	c := NewTTLCache()
	key := "acct|nonce1"
	if c.Used(key) {
		t.Fatalf("key should not be used initially")
	}
	c.Mark(key, 50*time.Millisecond)
	if !c.Used(key) {
		t.Fatalf("key should be marked as used within TTL")
	}
	time.Sleep(60 * time.Millisecond)
	if c.Used(key) {
		t.Fatalf("key should expire after TTL")
	}
}
