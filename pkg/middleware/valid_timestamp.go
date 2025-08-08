package middleware

import (
	"log/slog"
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
	// logger is used to record validation details.
	logger *slog.Logger
	// now returns the current time; useful to inject a deterministic clock in tests.
	now func() time.Time
}

// NewValidTimestamp constructs a ValidTimestamp with the provided inputs.
// Use a nil logger or now function to fall back to defaults.
func NewValidTimestamp(ts string, logger *slog.Logger, now func() time.Time) ValidTimestamp {
	return ValidTimestamp{ts: ts, logger: logger, now: now}
}

// Validate ensures the timestamp is numeric (Unix seconds) and within the allowed skew window.
// If disallowFuture is true, timestamps after "now" are rejected even if within positive skew.
func (v ValidTimestamp) Validate(skew time.Duration, disallowFuture bool) *http.ApiError {
	if v.ts == "" {
		if v.logger != nil {
			v.logger.Warn("missing timestamp")
		}
		return &http.ApiError{Message: "Invalid authentication headers", Status: 401}
	}

	epoch, err := strconv.ParseInt(v.ts, 10, 64)
	if err != nil {
		if v.logger != nil {
			v.logger.Warn("invalid timestamp format")
		}
		return &http.ApiError{Message: "Invalid authentication headers", Status: 401}
	}

	nowFn := v.now
	if nowFn == nil {
		nowFn = time.Now
	}
	now := nowFn().Unix()
	skewSecs := int64(skew.Seconds())
	minValue := now - skewSecs
	maxValue := now + skewSecs
	if disallowFuture {
		maxValue = now
	}

	if epoch < minValue {
		if v.logger != nil {
			v.logger.Warn("timestamp outside allowed window: too old")
		}
		return &http.ApiError{Message: "Invalid credentials", Status: 401}
	}

	if epoch > maxValue {
		if v.logger != nil {
			v.logger.Warn("timestamp outside allowed window: in the future")
		}
		return &http.ApiError{Message: "Invalid credentials", Status: 401}
	}

	return nil
}
