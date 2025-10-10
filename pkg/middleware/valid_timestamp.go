package middleware

import (
	"time"

	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/middleware/mwguards"
)

// ValidTimestamp re-exposes the mwguards timestamp helper for backwards compatibility.
type ValidTimestamp = mwguards.ValidTimestamp

// NewValidTimestamp mirrors the legacy constructor while delegating to the mwguards implementation.
func NewValidTimestamp(ts string, now func() time.Time) ValidTimestamp {
	return mwguards.NewValidTimestamp(ts, now)
}

// Validate validates the timestamp against the provided skew constraints.
func (v ValidTimestamp) Validate(skew time.Duration, disallowFuture bool) *http.ApiError {
	return mwguards.ValidTimestamp(v).Validate(skew, disallowFuture)
}
