package endpoint

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/getsentry/sentry-go"
)

func NewApiHandler(fn ApiHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			slog.Error("API Error", "message", err.Message, "status", err.Status)

			captureApiError(r, err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.Status)

			resp := ErrorResponse{
				Error:  err.Message,
				Status: err.Status,
				Data:   err.Data,
			}

			if result := json.NewEncoder(w).Encode(resp); result != nil {
				slog.Error("Could not encode error response", "error", result)
			}
		}
	}
}

func captureApiError(r *http.Request, apiErr *ApiError) {
	if apiErr == nil {
		return
	}

	errToCapture := error(apiErr)
	if apiErr.Err != nil {
		errToCapture = apiErr.Err
	}

	notify := func(hub *sentry.Hub) {
		hub.WithScope(func(scope *sentry.Scope) {
			scopeApiError := NewScopeApiError(scope, r, apiErr)

			scopeApiError.Enrich()

			// Set appropriate severity level based on status code
			// Authentication/authorization errors are logged as info for monitoring
			// without triggering alerts, while actual errors remain at error level
			level := getSentryLevel(apiErr.Status)
			scope.SetLevel(level)

			hub.CaptureException(errToCapture)
		})
	}

	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		notify(hub)
		return
	}

	notify(sentry.CurrentHub())
}

func getSentryLevel(status int) sentry.Level {
	// Expected client errors are logged as info for visibility without noise:
	// - 401 Unauthorized: Invalid credentials/tokens
	// - 403 Forbidden: Insufficient permissions
	// - 404 Not Found: Resource doesn't exist
	// - 429 Too Many Requests: Rate limiting
	switch status {
	case http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusTooManyRequests:
		return sentry.LevelInfo
	default:
		return sentry.LevelError
	}
}
