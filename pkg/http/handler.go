package http

import (
	"encoding/json"
	"log/slog"
	baseHttp "net/http"
	"strconv"

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

	level := sentry.LevelWarning
	if apiErr.Status >= baseHttp.StatusInternalServerError {
		level = sentry.LevelError
	}

	notify := func(hub *sentry.Hub) {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(level)
			scope.SetTag("http.method", r.Method)
			scope.SetTag("http.status_code", strconv.Itoa(apiErr.Status))
			scope.SetTag("http.route", r.URL.Path)
			scope.SetRequest(r)
			scope.SetExtra("api_error_status_text", baseHttp.StatusText(apiErr.Status))

			if apiErr.Data != nil {
				scope.SetExtra("api_error_data", apiErr.Data)
			}

			hub.CaptureException(apiErr)
		})
	}

	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		notify(hub)
		return
	}

	notify(sentry.CurrentHub())
}
