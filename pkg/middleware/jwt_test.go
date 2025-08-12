package middleware

import (
	baseHttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http"
)

func TestJWTMiddlewareHandle(t *testing.T) {
	handler, err := auth.MakeJWTHandler([]byte("mysecretjwtkey123"), time.Minute)
	if err != nil {
		t.Fatalf("make handler err: %v", err)
	}

	m := JWTMiddleware{Handler: handler}

	next := func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		claims, ok := GetJWTClaims(r.Context())
		if !ok {
			t.Fatalf("claims missing from context")
		}
		if claims.Username != "bob" {
			t.Fatalf("expected bob got %s", claims.Username)
		}
		return nil
	}

	token, err := handler.Generate("bob")
	if err != nil {
		t.Fatalf("generate token err: %v", err)
	}

	req := httptest.NewRequest(baseHttp.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	if err := m.Handle(next)(rr, req); err != nil {
		t.Fatalf("expected nil got %v", err)
	}
}

func TestJWTMiddlewareUnauthorized(t *testing.T) {
	handler, err := auth.MakeJWTHandler([]byte("anotherjwtsecret"), time.Minute)
	if err != nil {
		t.Fatalf("make handler err: %v", err)
	}

	m := JWTMiddleware{Handler: handler}

	req := httptest.NewRequest(baseHttp.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	errResp := m.Handle(func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError { return nil })(rr, req)
	if errResp == nil {
		t.Fatalf("expected error for missing token")
	}
	if errResp.Status != baseHttp.StatusUnauthorized {
		t.Fatalf("expected unauthorized got %d", errResp.Status)
	}
}
