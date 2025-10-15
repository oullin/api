package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	baseHttp "net/http"
	"strconv"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/oullin/pkg/portal"
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

	errToCapture := error(apiErr)
	if apiErr.Err != nil {
		errToCapture = apiErr.Err
	}

	notify := func(hub *sentry.Hub) {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(level)
			scope.SetTag("http.method", r.Method)
			scope.SetTag("http.status_code", strconv.Itoa(apiErr.Status))
			scope.SetTag("http.route", r.URL.Path)
			if requestID := requestIDFrom(r); requestID != "" {
				scope.SetTag("http.request_id", requestID)
				scope.SetExtra("http_request_id", requestID)
			}
			scope.SetRequest(r)
			scope.SetExtra("api_error_status_text", baseHttp.StatusText(apiErr.Status))
			scope.SetExtra("api_error_message", apiErr.Message)

			enrichScopeWithApiError(scope, r, apiErr)

			hub.CaptureException(errToCapture)
		})
	}

	if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
		notify(hub)
		return
	}

	notify(sentry.CurrentHub())
}

func enrichScopeWithApiError(scope *sentry.Scope, r *baseHttp.Request, apiErr *ApiError) {
	if apiErr.Data != nil {
		scope.SetExtra("api_error_data", apiErr.Data)
	}

	if apiErr.Err != nil {
		scope.SetExtra("api_error_cause", apiErr.Err.Error())
		scope.SetTag("api.error.cause_type", fmt.Sprintf("%T", apiErr.Err))

		if chain := buildErrorChain(apiErr.Err); len(chain) > 0 {
			scope.SetExtra("api_error_cause_chain", chain)
		}
	}

	if accountName := accountNameFrom(r); accountName != "" {
		scope.SetExtra("api_account_name", accountName)
	}

	if username := headerValue(r, portal.UsernameHeader); username != "" {
		scope.SetExtra("api_username_header", username)
	}

	if origin := headerValue(r, portal.IntendedOriginHeader); origin != "" {
		scope.SetExtra("api_intended_origin", origin)
	}

	if ts := headerValue(r, portal.TimestampHeader); ts != "" {
		scope.SetExtra("api_request_timestamp", ts)
	}

	if nonce := headerValue(r, portal.NonceHeader); nonce != "" {
		scope.SetExtra("api_request_nonce", nonce)
	}

	if publicKey := headerValue(r, portal.TokenHeader); publicKey != "" {
		scope.SetExtra("api_public_key", publicKey)
	}

	if clientIP := strings.TrimSpace(portal.ParseClientIP(r)); clientIP != "" {
		scope.SetExtra("http_client_ip", clientIP)
	}
}

func requestIDFrom(r *baseHttp.Request) string {
	if v, ok := r.Context().Value(portal.RequestIDKey).(string); ok {
		if id := strings.TrimSpace(v); id != "" {
			return id
		}
	}

	return headerValue(r, portal.RequestIDHeader)
}

func accountNameFrom(r *baseHttp.Request) string {
	if v, ok := r.Context().Value(portal.AuthAccountNameKey).(string); ok {
		if name := strings.TrimSpace(v); name != "" {
			return name
		}
	}

	return headerValue(r, portal.UsernameHeader)
}

func headerValue(r *baseHttp.Request, key string) string {
	return strings.TrimSpace(r.Header.Get(key))
}

func buildErrorChain(err error) []string {
	chain := make([]string, 0, 4)

	for current := err; current != nil; current = errors.Unwrap(current) {
		chain = append(chain, current.Error())
	}

	return chain
}
