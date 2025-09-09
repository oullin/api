package kernel

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
)

func TestPingRoute_PublicMiddleware(t *testing.T) {
	fixedTime := time.Unix(1700000000, 0)
	pm := middleware.MakePublicMiddleware("", false)
	rv := reflect.ValueOf(&pm).Elem().FieldByName("now")
	reflect.NewAt(rv.Type(), unsafe.Pointer(rv.UnsafeAddr())).Elem().Set(reflect.ValueOf(func() time.Time { return fixedTime }))

	r := Router{
		Mux: http.NewServeMux(),
		Pipeline: middleware.Pipeline{
			PublicMiddleware: pm,
		},
	}
	r.Ping()

	t.Run("request without public headers is unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping", nil)
		rec := httptest.NewRecorder()
		r.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("request with public headers succeeds", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping", nil)
		req.Header.Set(portal.RequestIDHeader, "req-1")
		req.Header.Set(portal.TimestampHeader, fmt.Sprintf("%d", fixedTime.Unix()))
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		rec := httptest.NewRecorder()
		r.Mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
