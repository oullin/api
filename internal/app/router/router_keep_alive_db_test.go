package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	env "github.com/oullin/internal/app/config"
	"github.com/oullin/internal/app/middleware"
	"github.com/oullin/internal/app/router"
	"github.com/oullin/internal/testutil/dbtest"
)

func TestKeepAliveDBRoute(t *testing.T) {
	db, _ := dbtest.NewTestDB(t)
	r := router.Router{
		Env:      &env.Environment{Ping: env.PingEnvironment{Username: "user", Password: "pass"}},
		Db:       db,
		Mux:      http.NewServeMux(),
		Pipeline: middleware.Pipeline{PublicMiddleware: middleware.NewPublicMiddleware("", false)},
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
