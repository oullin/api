package endpoint_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/pkg/endpoint"
)

func TestResponse_RespondOkAndHasCache(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	r := endpoint.NewResponseFrom("salt", rec, req)

	if err := r.RespondOk(map[string]string{"a": "b"}); err != nil {
		t.Fatalf("respond ok: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("ETag") == "" || rec.Header().Get("Cache-Control") == "" {
		t.Fatalf("expected cache headers to be set")
	}

	req.Header.Set("If-None-Match", rec.Header().Get("ETag"))

	if !r.HasCache() {
		t.Fatalf("expected cache to be detected")
	}
}

func TestResponse_NoCache(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	r := endpoint.NewNoCacheResponse(rec, req)

	if err := r.RespondOk(map[string]string{"a": "b"}); err != nil {
		t.Fatalf("respond ok: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected cache-control 'no-store', got %q", rec.Header().Get("Cache-Control"))
	}

	if rec.Header().Get("ETag") != "" {
		t.Fatalf("expected empty etag for no-cache response")
	}

	if r.HasCache() {
		t.Fatalf("expected no cache")
	}
}

func TestResponse_WithHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r := endpoint.NewResponseFrom("salt", rec, req)
	called := false

	r.WithHeaders(func(w http.ResponseWriter) { called = true })

	if !called {
		t.Fatalf("expected callback to be called")
	}
}

func TestResponse_NotModified(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	r := endpoint.NewResponseFrom("salt", rec, req)
	r.RespondWithNotModified()

	if rec.Code != http.StatusNotModified {
		t.Fatalf("expected status %d, got %d", http.StatusNotModified, rec.Code)
	}
}

func TestApiErrorHelpers(t *testing.T) {
	if endpoint.InternalError("x").Status != http.StatusInternalServerError {
		t.Fatalf("expected internal error status %d", http.StatusInternalServerError)
	}

	if endpoint.BadRequestError("x").Status != http.StatusBadRequest {
		t.Fatalf("expected bad request status %d", http.StatusBadRequest)
	}

	if endpoint.NotFound("x").Status != http.StatusNotFound {
		t.Fatalf("expected not found status %d", http.StatusNotFound)
	}
}
