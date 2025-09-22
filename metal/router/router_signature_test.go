package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
)

func TestSignatureRoute_PublicMiddleware(t *testing.T) {
	r := Router{
		Mux: http.NewServeMux(),
		Pipeline: middleware.Pipeline{
			PublicMiddleware: middleware.MakePublicMiddleware("", false),
		},
		validator: portal.GetDefaultValidator(),
	}
	r.Signature()

	t.Run("request without public headers is unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/generate-signature", nil)
		rec := httptest.NewRecorder()
		r.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("request with public headers but invalid body is bad request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/generate-signature", strings.NewReader("{"))
		req.Header.Set(portal.RequestIDHeader, "req-1")
		req.Header.Set(portal.TimestampHeader, fmt.Sprintf("%d", time.Now().Unix()))
		rec := httptest.NewRecorder()
		r.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})
}
