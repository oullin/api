package middleware

import "github.com/oullin/pkg/middleware/mwguards"

type ValidTimestamp = mwguards.ValidTimestamp

var (
	NewValidTimestamp    = mwguards.NewValidTimestamp
	TimestampTooOldError = mwguards.TimestampTooOldError
	TimestampTooNewError = mwguards.TimestampTooNewError
)
