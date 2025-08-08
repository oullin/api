package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	pkgHttp "github.com/oullin/pkg/http"
)

func TestTokenMiddlewareErrors(t *testing.T) {
	tm := TokenCheckMiddleware{}

	e := tm.getInvalidRequestError()

	if e.Status != 401 || e.Message == "" {
		t.Fatalf("invalid request error")
	}

	e = tm.getInvalidTokenFormatError()

	if e.Status != 401 {
		t.Fatalf("invalid token error")
	}

	e = tm.getUnauthenticatedError()

	if e.Status != 401 {
		t.Fatalf("unauthenticated error")
	}
}

func TestTokenMiddlewareHandleInvalid(t *testing.T) {
	tm := MakeTokenMiddleware(nil, nil)

	handler := tm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
		return nil
	})

	rec := httptest.NewRecorder()
	err := handler(rec, httptest.NewRequest("GET", "/", nil))

	if err == nil || err.Status != 401 {
		t.Fatalf("expected unauthorized")
	}
}
