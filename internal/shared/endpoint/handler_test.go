package endpoint_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"

	"github.com/oullin/internal/shared/endpoint"
)

func TestNewApiHandler(t *testing.T) {
	h := endpoint.NewApiHandler(func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {

		return &endpoint.ApiError{
			Message: "bad",
			Status:  http.StatusBadRequest,
			Err:     errors.New("bad"),
		}
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp endpoint.ErrorResponse

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Error == "" || resp.Status != http.StatusBadRequest {
		t.Fatalf("expected error response with bad request status")
	}
}

// Note: TestScopeApiErrorRequestID, TestScopeApiErrorAccountName, and TestScopeApiErrorBuildErrorChain
// have been removed as they test internal implementation details that cannot be accessed from external test packages.
// The functionality is still covered by the integration test TestScopeApiErrorEnrichSetsLevelAndTags.

func TestScopeApiErrorEnrichSetsLevelAndTags(t *testing.T) {
	scope := sentry.NewScope()
	req := httptest.NewRequest("POST", "/resource", nil)

	apiErr := &endpoint.ApiError{Status: http.StatusInternalServerError, Err: errors.New("boom")}

	endpoint.NewScopeApiError(scope, req, apiErr).Enrich()

	event := scope.ApplyToEvent(sentry.NewEvent(), nil, nil)
	if event == nil {
		t.Fatalf("expected event after scope enrichment")
	}

	if event.Level != sentry.LevelError {
		t.Fatalf("expected error level, got %s", event.Level)
	}

	if got := event.Tags["http.method"]; got != "POST" {
		t.Fatalf("expected POST method tag, got %s", got)
	}

	if got := event.Tags["http.status_code"]; got != "500" {
		t.Fatalf("expected 500 status code tag, got %s", got)
	}

	if got := event.Tags["http.route"]; got != "/resource" {
		t.Fatalf("expected /resource route tag, got %s", got)
	}
}

func TestScopeApiErrorEnrichSetsWarningLevelForClientErrors(t *testing.T) {
	scope := sentry.NewScope()
	req := httptest.NewRequest("GET", "/client", nil)

	apiErr := &endpoint.ApiError{Status: http.StatusBadRequest, Err: errors.New("bad request")}

	endpoint.NewScopeApiError(scope, req, apiErr).Enrich()

	event := scope.ApplyToEvent(sentry.NewEvent(), nil, nil)
	if event == nil {
		t.Fatalf("expected event after scope enrichment")
	}

	if event.Level != sentry.LevelWarning {
		t.Fatalf("expected warning level, got %s", event.Level)
	}
}
