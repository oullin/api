package endpoint

import (
	"encoding/json"
	"log/slog"
	baseHttp "net/http"

	"github.com/getsentry/sentry-go"
)

func MakeApiHandler(fn ApiHandler) baseHttp.HandlerFunc {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) {
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

func captureApiError(r *baseHttp.Request, apiErr *ApiError) {
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

			hub.CaptureException(errToCapture)
		})
	}

	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		notify(hub)
		return
	}

	notify(sentry.CurrentHub())
}
