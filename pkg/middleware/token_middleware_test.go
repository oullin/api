package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgHttp "github.com/oullin/pkg/http"
)

func TestTokenMiddlewareErrors(t *testing.T) {
	tm := TokenCheckMiddleware{}

	e := tm.getInvalidRequestError()

	if e.Status != http.StatusUnauthorized || e.Message == "" {
		t.Fatalf("invalid request error")
	}

	e = tm.getInvalidTokenFormatError()

	if e.Status != http.StatusUnauthorized {
		t.Fatalf("invalid token error")
	}

	e = tm.getUnauthenticatedError()

	if e.Status != http.StatusUnauthorized {
		t.Fatalf("unauthenticated error")
	}
}

func TestTokenMiddlewareHandle_RequiresRequestID(t *testing.T) {
	tm := MakeTokenMiddleware(nil, nil)

	handler := tm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	// No X-Request-ID present
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected 401 when X-Request-ID is missing, got %#v", err)
	}
}

func TestTokenMiddlewareHandleInvalid(t *testing.T) {
	tm := MakeTokenMiddleware(nil, nil)

	handler := tm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-1")
	// Missing other auth headers triggers invalid request
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for missing auth headers, got %#v", err)
	}
}

func TestValidateAndGetHeaders_MissingAndInvalidFormat(t *testing.T) {
	tm := MakeTokenMiddleware(nil, nil)
	logger := slogNoop()
	req := httptest.NewRequest("GET", "/", nil)
	// All empty
	if _, _, _, _, _, apiErr := tm.validateAndGetHeaders(req, logger); apiErr == nil || apiErr.Status != http.StatusUnauthorized {
		t.Fatalf("expected error for missing headers")
	}

	// Set minimal headers but invalid token format (not pk_/sk_ prefix or too short)
	req.Header.Set("X-API-Username", "alice")
	req.Header.Set("X-API-Key", "badtoken")
	req.Header.Set("X-API-Signature", "sig")
	req.Header.Set("X-API-Timestamp", "1700000000")
	req.Header.Set("X-API-Nonce", "n1")
	if _, _, _, _, _, apiErr := tm.validateAndGetHeaders(req, logger); apiErr == nil || apiErr.Status != http.StatusUnauthorized {
		t.Fatalf("expected error for invalid token format")
	}
}

func TestReadBodyHash_RestoresBody(t *testing.T) {
	tm := MakeTokenMiddleware(nil, nil)
	logger := slogNoop()
	body := "{\"a\":1}"
	req := httptest.NewRequest("POST", "/x", bytes.NewBufferString(body))
	hash, apiErr := tm.readBodyHash(req, logger)
	if apiErr != nil || hash == "" {
		t.Fatalf("expected body hash, got err=%v hash=%q", apiErr, hash)
	}
	// Now the body should be readable again for downstream
	b, _ := io.ReadAll(req.Body)
	if string(b) != body {
		t.Fatalf("expected body to be restored, got %q", string(b))
	}
}

func TestAttachContext(t *testing.T) {
	tm := MakeTokenMiddleware(nil, nil)
	req := httptest.NewRequest("GET", "/", nil)
	r := tm.attachContext(req, "Alice", "RID-123")
	if r == req {
		t.Fatalf("expected a new request with updated context")
	}
	if r.Context() == nil {
		t.Fatalf("expected non-nil context")
	}
}
