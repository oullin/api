package endpoint

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponse_RespondOkAndHasCache(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	r := NewResponseFrom("salt", rec, req)

	if err := r.RespondOk(map[string]string{"a": "b"}); err != nil {
		t.Fatalf("respond: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	if rec.Header().Get("ETag") == "" || rec.Header().Get("Cache-Control") == "" {
		t.Fatalf("headers missing")
	}

	req.Header.Set("If-None-Match", r.etag)

	if !r.HasCache() {
		t.Fatalf("expected cache")
	}
}

func TestResponse_NoCache(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	r := NewNoCacheResponse(rec, req)

	if err := r.RespondOk(map[string]string{"a": "b"}); err != nil {
		t.Fatalf("respond: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected cache-control: %s", rec.Header().Get("Cache-Control"))
	}

	if rec.Header().Get("ETag") != "" {
		t.Fatalf("etag should be empty")
	}

	if r.HasCache() {
		t.Fatalf("expected no cache")
	}
}

func TestResponse_WithHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	r := NewResponseFrom("salt", rec, req)
	called := false

	r.WithHeaders(func(w http.ResponseWriter) { called = true })

	if !called {
		t.Fatalf("callback not called")
	}
}

func TestResponse_NotModified(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	r := NewResponseFrom("salt", rec, req)
	r.RespondWithNotModified()

	if rec.Code != http.StatusNotModified {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestApiErrorHelpers(t *testing.T) {
	if InternalError("x").Status != http.StatusInternalServerError {
		t.Fatalf("internal status")
	}

	if BadRequestError("x").Status != http.StatusBadRequest {
		t.Fatalf("bad req status")
	}

	if NotFound("x").Status != http.StatusNotFound {
		t.Fatalf("not found")
	}
}
