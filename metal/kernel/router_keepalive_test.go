package kernel

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
)

func TestKeepAliveRoute_PublicMiddleware(t *testing.T) {
	r := Router{
		Mux: http.NewServeMux(),
		Pipeline: middleware.Pipeline{
			PublicMiddleware: middleware.MakePublicMiddleware("", false),
		},
	}
	r.KeepAlive()

	t.Run("request without public headers is unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/keep-alive", nil)
		rec := httptest.NewRecorder()
		r.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("request with public headers succeeds", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/keep-alive", nil)
		req.Header.Set(portal.RequestIDHeader, "req-1")
		req.Header.Set(portal.TimestampHeader, fmt.Sprintf("%d", time.Now().Unix()))
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		rec := httptest.NewRecorder()
		r.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
