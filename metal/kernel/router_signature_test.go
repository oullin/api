package kernel

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
		Mux:       http.NewServeMux(),
		Pipeline:  middleware.Pipeline{},
		validator: portal.GetDefaultValidator(),
	}
	r.Signature()

	req := httptest.NewRequest("POST", "/generate-signature", nil)
	rec := httptest.NewRecorder()
	r.Mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	req2 := httptest.NewRequest("POST", "/generate-signature", strings.NewReader("{"))
	req2.Header.Set(portal.RequestIDHeader, "req-1")
	req2.Header.Set(portal.TimestampHeader, fmt.Sprintf("%d", time.Now().Unix()))
	rec2 := httptest.NewRecorder()
	r.Mux.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec2.Code)
	}
}
