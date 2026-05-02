package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/middleware"
)

func TestHealthRoute(t *testing.T) {
	r := router.Router{
		Mux:      http.NewServeMux(),
		Pipeline: middleware.Pipeline{PublicMiddleware: middleware.NewPublicMiddleware("", false)},
	}
	r.Health()

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	r.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
