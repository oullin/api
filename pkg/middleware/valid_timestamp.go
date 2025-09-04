package middleware

import (
	baseHttp "net/http"
	"strconv"
	"time"

	"github.com/oullin/pkg/http"
)

// ValidTimestamp encapsulates timestamp validation context.
// It accepts: the raw timestamp string (ts), a logger, and a clock (now) function.
// Use Validate to check against a provided skew window and future-time policy.
type ValidTimestamp struct {
	// ts is the timestamp string (expected Unix epoch in seconds).
	ts string

	// now returns the current time; useful to inject a deterministic clock in tests.
	now func() time.Time
}

func NewValidTimestamp(ts string, now func() time.Time) ValidTimestamp {
	return ValidTimestamp{
		ts:  ts,
		now: now,
	}
}

func (v ValidTimestamp) Validate(skew time.Duration, disallowFuture bool) *http.ApiError {
	if v.ts == "" {
		return &http.ApiError{Message: "Invalid authentication headers", Status: baseHttp.StatusUnauthorized}
	}

	epoch, err := strconv.ParseInt(v.ts, 10, 64)
	if err != nil {
		return &http.ApiError{Message: "Invalid authentication headers", Status: baseHttp.StatusUnauthorized}
	}

	nowFn := v.now
	if nowFn == nil {
		nowFn = time.Now
	}

	now := nowFn().Unix()
	if skew < 0 {
		skew = -skew
	}

	skewSecs := int64(skew / time.Second)
	minValue := now - skewSecs
	maxValue := now + skewSecs

	if disallowFuture {
		maxValue = now
	}

	if epoch < minValue {
		return &http.ApiError{Message: "Request timestamp expired", Status: baseHttp.StatusUnauthorized}
	}

	if epoch > maxValue {
		return &http.ApiError{Message: "Request timestamp invalid", Status: baseHttp.StatusUnauthorized}
	}

	return nil
}
