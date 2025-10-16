package mwguards

import (
	"net/http"
	"strconv"
	"time"

	"github.com/oullin/pkg/endpoint"
)

type ValidTimestamp struct {
	ts  string
	now func() time.Time
}

func NewValidTimestamp(ts string, now func() time.Time) ValidTimestamp {
	return ValidTimestamp{
		ts:  ts,
		now: now,
	}
}

func (v ValidTimestamp) Validate(skew time.Duration, disallowFuture bool) *endpoint.ApiError {
	if v.ts == "" {
		return &endpoint.ApiError{Message: "Invalid authentication headers", Status: http.StatusUnauthorized}
	}

	epoch, err := strconv.ParseInt(v.ts, 10, 64)
	if err != nil {
		return &endpoint.ApiError{Message: "Invalid authentication headers", Status: http.StatusUnauthorized}
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
		return TimestampTooOldError("Request timestamp is too old or invalid", "Request timestamp invalid")
	}

	if epoch > maxValue {
		return TimestampTooNewError("Request timestamp is too recent or invalid", "Request timestamp invalid")
	}

	return nil
}
