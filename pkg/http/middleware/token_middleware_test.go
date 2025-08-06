package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	pkgAuth "github.com/oullin/pkg/auth"
	pkgHttp "github.com/oullin/pkg/http"
)

func TestTokenMiddlewareErrors(t *testing.T) {
	tm := TokenCheckMiddleware{}

	e := tm.getInvalidRequestError("a", "b", "c")
	if e.Status != 403 || e.Message == "" {
		t.Fatalf("invalid request error")
	}
	e = tm.getInvalidTokenFormatError("pk_x", pkgAuth.ValidateTokenFormat("bad"))
	if e.Status != 403 {
		t.Fatalf("invalid token error")
	}
	e = tm.getUnauthenticatedError("a", "b", "c")
	if e.Status != 403 {
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

	if err == nil || err.Status != 403 {
		t.Fatalf("expected forbidden")
	}
}
