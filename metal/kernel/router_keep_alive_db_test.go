package kernel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	handlertests "github.com/oullin/handler/tests"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/middleware"
)

func TestKeepAliveDBRoute(t *testing.T) {
	db, _ := handlertests.MakeTestDB(t)
	r := Router{
		Env:      &env.Environment{Ping: env.PingEnvironment{Username: "user", Password: "pass"}},
		Db:       db,
		Mux:      http.NewServeMux(),
		Pipeline: middleware.Pipeline{PublicMiddleware: middleware.MakePublicMiddleware("", false)},
	}
	r.KeepAliveDB()

	t.Run("valid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping-db", nil)
		req.SetBasicAuth("user", "pass")
		rec := httptest.NewRecorder()
		r.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping-db", nil)
		req.SetBasicAuth("bad", "creds")
		rec := httptest.NewRecorder()
		r.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})
}
