package mwguards

import (
	baseHttp "net/http"
	"strconv"
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time { return func() time.Time { return t } }

func TestNewValidTimestampConstructor(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	vt := NewValidTimestamp("123", fixedClock(base))

	if vt.ts != "123" {
		t.Fatalf("expected ts to be set by constructor")
	}
	if vt.now == nil || vt.now().Unix() != base.Unix() {
		t.Fatalf("expected now clock to be set by constructor")
	}
}

func TestValidate_EmptyTimestamp(t *testing.T) {
	vt := NewValidTimestamp("", fixedClock(time.Unix(1_700_000_000, 0)))
	err := vt.Validate(5*time.Minute, false)
	if err == nil || err.Status != baseHttp.StatusUnauthorized || err.Message != "Invalid authentication headers" {
		t.Fatalf("expected invalid request error for empty timestamp, got %#v", err)
	}
}

func TestValidate_NonNumericTimestamp(t *testing.T) {
	vt := NewValidTimestamp("abc", fixedClock(time.Unix(1_700_000_000, 0)))
	err := vt.Validate(5*time.Minute, false)
	if err == nil || err.Status != baseHttp.StatusUnauthorized || err.Message != "Invalid authentication headers" {
		t.Fatalf("expected invalid request error for non-numeric timestamp, got %#v", err)
	}
}

func TestValidate_TooOldTimestamp(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	skew := 60 * time.Second
	oldTs := strconv.FormatInt(base.Add(-skew).Add(-1*time.Second).Unix(), 10)
	vt := NewValidTimestamp(oldTs, fixedClock(base))
	err := vt.Validate(skew, false)
	if err == nil || err.Status != baseHttp.StatusUnauthorized || err.Message != "Request timestamp expired" {
		t.Fatalf("expected unauthenticated for too old timestamp, got %#v", err)
	}
}

func TestValidate_FutureWithinSkew_Behavior(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	skew := 60 * time.Second
	futureWithin := strconv.FormatInt(base.Add(30*time.Second).Unix(), 10)

	// Allowed when disallowFuture=false
	vt := NewValidTimestamp(futureWithin, fixedClock(base))
	if err := vt.Validate(skew, false); err != nil {
		t.Fatalf("expected future timestamp within skew to be allowed when disallowFuture=false, got %#v", err)
	}

	// Rejected when disallowFuture=true
	vt = NewValidTimestamp(futureWithin, fixedClock(base))
	err := vt.Validate(skew, true)
	if err == nil || err.Status != baseHttp.StatusUnauthorized || err.Message != "Request timestamp invalid" {
		t.Fatalf("expected unauthenticated for future timestamp when disallowFuture=true, got %#v", err)
	}
}

func TestValidate_Boundaries(t *testing.T) {
	base := time.Unix(1_700_000_000, 0)
	skew := 60 * time.Second
	minExact := strconv.FormatInt(base.Add(-skew).Unix(), 10)
	maxExact := strconv.FormatInt(base.Add(skew).Unix(), 10)
	nowExact := strconv.FormatInt(base.Unix(), 10)

	// Lower boundary inclusive
	vt := NewValidTimestamp(minExact, fixedClock(base))
	if err := vt.Validate(skew, false); err != nil {
		t.Fatalf("expected min boundary to pass, got %#v", err)
	}

	// Upper boundary inclusive when disallowFuture=false
	vt = NewValidTimestamp(maxExact, fixedClock(base))
	if err := vt.Validate(skew, false); err != nil {
		t.Fatalf("expected max boundary to pass when disallowFuture=false, got %#v", err)
	}

	// When disallowFuture=true, upper boundary becomes 'now'
	vt = NewValidTimestamp(nowExact, fixedClock(base))
	if err := vt.Validate(skew, true); err != nil {
		t.Fatalf("expected 'now' to pass when disallowFuture=true, got %#v", err)
	}
}
