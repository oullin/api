package http

import (
	"errors"
	"fmt"
	baseHttp "net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/oullin/pkg/portal"
)

type ScopeApiError struct {
	scope   *sentry.Scope
	request *baseHttp.Request
	apiErr  *ApiError
}

func NewScopeApiError(scope *sentry.Scope, r *baseHttp.Request, apiErr *ApiError) *ScopeApiError {
	return &ScopeApiError{scope: scope, request: r, apiErr: apiErr}
}

func (s *ScopeApiError) RequestID() string {
	if s == nil || s.request == nil {
		return ""
	}

	if v, ok := s.request.Context().Value(portal.RequestIDKey).(string); ok {
		if id := strings.TrimSpace(v); id != "" {
			return id
		}
	}

	return s.headerValue(portal.RequestIDHeader)
}

func (s *ScopeApiError) Enrich() {
	if s == nil || s.scope == nil || s.request == nil || s.apiErr == nil {
		return
	}

	s.scope.SetRequest(s.request)
	s.scope.SetExtra("api_error_status_text", baseHttp.StatusText(s.apiErr.Status))
	s.scope.SetExtra("api_error_message", s.apiErr.Message)

	if requestID := s.RequestID(); requestID != "" {
		s.scope.SetTag("http.request_id", requestID)
		s.scope.SetExtra("http_request_id", requestID)
	}

	if s.apiErr.Data != nil {
		s.scope.SetExtra("api_error_data", s.apiErr.Data)
	}

	if s.apiErr.Err != nil {
		s.scope.SetExtra("api_error_cause", s.apiErr.Err.Error())
		s.scope.SetTag("api.error.cause_type", fmt.Sprintf("%T", s.apiErr.Err))

		s.scope.SetExtra("api_error_cause_chain", s.buildErrorChain(s.apiErr.Err))
	}

	if accountName := s.accountName(); accountName != "" {
		s.scope.SetExtra("api_account_name", accountName)
	}

	if username := s.headerValue(portal.UsernameHeader); username != "" {
		s.scope.SetExtra("api_username_header", username)
	}

	if origin := s.headerValue(portal.IntendedOriginHeader); origin != "" {
		s.scope.SetExtra("api_intended_origin", origin)
	}

	if ts := s.headerValue(portal.TimestampHeader); ts != "" {
		s.scope.SetExtra("api_request_timestamp", ts)
	}

	if nonce := s.headerValue(portal.NonceHeader); nonce != "" {
		s.scope.SetExtra("api_request_nonce", nonce)
	}

	if publicKey := s.headerValue(portal.TokenHeader); publicKey != "" {
		s.scope.SetExtra("api_public_key", publicKey)
	}

	if clientIP := strings.TrimSpace(portal.ParseClientIP(s.request)); clientIP != "" {
		s.scope.SetExtra("http_client_ip", clientIP)
	}
}

func (s *ScopeApiError) accountName() string {
	if s == nil || s.request == nil {
		return ""
	}

	if v, ok := s.request.Context().Value(portal.AuthAccountNameKey).(string); ok {
		if name := strings.TrimSpace(v); name != "" {
			return name
		}
	}

	return s.headerValue(portal.UsernameHeader)
}

func (s *ScopeApiError) headerValue(key string) string {
	if s == nil || s.request == nil {
		return ""
	}

	return strings.TrimSpace(s.request.Header.Get(key))
}

func (s *ScopeApiError) buildErrorChain(err error) []string {
	chain := make([]string, 0, 4)

	for current := err; current != nil; current = errors.Unwrap(current) {
		chain = append(chain, current.Error())
	}

	return chain
}
