package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/pkg/portal"
)

func TestMakeApiHandler(t *testing.T) {
	h := MakeApiHandler(func(w http.ResponseWriter, r *http.Request) *ApiError {

		return &ApiError{
			Message: "bad",
			Status:  http.StatusBadRequest,
		}
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}

	var resp ErrorResponse

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Error == "" || resp.Status != http.StatusBadRequest {
		t.Fatalf("invalid response")
	}
}

func TestScopeApiErrorRequestID(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(portal.RequestIDHeader, "header-id")

	scopeApiError := &ScopeApiError{request: req}

	if got := scopeApiError.RequestID(); got != "header-id" {
		t.Fatalf("expected header request id, got %s", got)
	}

	ctxReq := req.WithContext(context.WithValue(req.Context(), portal.RequestIDKey, "context-id"))

	scopeApiError.request = ctxReq

	if got := scopeApiError.RequestID(); got != "context-id" {
		t.Fatalf("expected context request id, got %s", got)
	}
}

func TestScopeApiErrorAccountName(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set(portal.UsernameHeader, "header-user")

	scopeApiError := &ScopeApiError{request: req}

	if got := scopeApiError.accountName(); got != "header-user" {
		t.Fatalf("expected header user, got %s", got)
	}

	ctxReq := req.WithContext(context.WithValue(req.Context(), portal.AuthAccountNameKey, "context-user"))

	scopeApiError.request = ctxReq

	if got := scopeApiError.accountName(); got != "context-user" {
		t.Fatalf("expected context user, got %s", got)
	}
}

func TestScopeApiErrorBuildErrorChain(t *testing.T) {
	root := errors.New("root")
	wrapped := fmt.Errorf("layer: %w", root)

	chain := (&ScopeApiError{}).buildErrorChain(wrapped)

	if len(chain) != 2 {
		t.Fatalf("expected 2 errors in chain, got %d", len(chain))
	}

	if chain[0] != wrapped.Error() || chain[1] != root.Error() {
		t.Fatalf("unexpected error chain: %#v", chain)
	}
}
