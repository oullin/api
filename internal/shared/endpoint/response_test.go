package endpoint_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/oullin/internal/shared/endpoint"
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

func TestResponse_FromPayloadUsesContentHash(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	type payload struct {
		Version string `json:"version"`
		Data    []struct {
			Title string `json:"title"`
		} `json:"data"`
	}
	first := payload{Version: "v1", Data: []struct {
		Title string `json:"title"`
	}{{Title: "first"}}}

	r, err := endpoint.NewResponseFromPayload(first, 3600, rec, req)
	if err != nil {
		t.Fatalf("response from payload: %v", err)
	}

	if err := r.RespondOk(first); err != nil {
		t.Fatalf("respond ok: %v", err)
	}

	second := payload{Version: "v1", Data: []struct {
		Title string `json:"title"`
	}{{Title: "second"}}}
	otherRec := httptest.NewRecorder()
	otherResp, err := endpoint.NewResponseFromPayload(second, 3600, otherRec, req)
	if err != nil {
		t.Fatalf("other response from payload: %v", err)
	}

	if err := otherResp.RespondOk(second); err != nil {
		t.Fatalf("other respond ok: %v", err)
	}

	if rec.Header().Get("ETag") == "" {
		t.Fatalf("expected etag to be set")
	}

	if rec.Header().Get("ETag") == otherRec.Header().Get("ETag") {
		t.Fatalf("expected etag to change when payload content changes")
	}

	req.Header.Set("If-None-Match", rec.Header().Get("ETag"))

	if !r.HasCache() {
		t.Fatalf("expected payload etag to satisfy conditional request")
	}
}

type marshalCounter struct {
	calls atomic.Int32
}

func (m *marshalCounter) MarshalJSON() ([]byte, error) {
	m.calls.Add(1)

	return []byte(`{"version":"v1"}`), nil
}

func TestResponse_FromPayloadReusesMarshaledBody(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	payload := &marshalCounter{}

	r, err := endpoint.NewResponseFromPayload(payload, 3600, rec, req)
	if err != nil {
		t.Fatalf("response from payload: %v", err)
	}

	if err := r.RespondOk(payload); err != nil {
		t.Fatalf("respond ok: %v", err)
	}

	if got := payload.calls.Load(); got != 1 {
		t.Fatalf("expected one marshal call, got %d", got)
	}
}

func TestResponse_ForPayloadDisablesCache(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	r, err := endpoint.NewResponseForPayload(map[string]string{"a": "b"}, 3600, false, rec, req)
	if err != nil {
		t.Fatalf("response for payload: %v", err)
	}

	if err := r.RespondOk(map[string]string{"a": "b"}); err != nil {
		t.Fatalf("respond ok: %v", err)
	}

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store cache-control, got %q", rec.Header().Get("Cache-Control"))
	}

	if rec.Header().Get("ETag") != "" {
		t.Fatalf("expected empty etag for disabled cache")
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

func TestNewResponseFromPayload_RejectsOversizedPayload(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	oversized := map[string]string{
		"data": string(make([]byte, endpoint.MaxResponseCacheSize+1)),
	}

	_, err := endpoint.NewResponseFromPayload(oversized, 3600, rec, req)
	if !errors.Is(err, endpoint.ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestNewResponseForPayload_FallsBackOnOversizedPayload(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	oversized := map[string]string{
		"data": string(make([]byte, endpoint.MaxResponseCacheSize+1)),
	}

	r, err := endpoint.NewResponseForPayload(oversized, 3600, true, rec, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := r.RespondOk(oversized); err != nil {
		t.Fatalf("respond ok: %v", err)
	}

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store cache-control, got %q", rec.Header().Get("Cache-Control"))
	}

	if rec.Header().Get("ETag") != "" {
		t.Fatalf("expected empty etag for oversized fallback, got %q", rec.Header().Get("ETag"))
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
