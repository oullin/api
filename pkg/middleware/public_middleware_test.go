package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	pkgHttp "github.com/oullin/pkg/http"
	"github.com/oullin/pkg/limiter"
)

func TestPublicMiddleware_InvalidHeaders(t *testing.T) {
	pm := MakePublicMiddleware()
	handler := pm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-1")
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for missing timestamp, got %#v", err)
	}
}

func TestPublicMiddleware_TimestampExpired(t *testing.T) {
	pm := MakePublicMiddleware()
	base := time.Unix(1_700_000_000, 0)
	pm.now = func() time.Time { return base }
	handler := pm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-1")
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	old := base.Add(-10 * time.Minute).Unix()
	req.Header.Set("X-API-Timestamp", strconv.FormatInt(old, 10))
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for old timestamp, got %#v", err)
	}
}

func TestPublicMiddleware_RateLimitAndReplay(t *testing.T) {
	pm := MakePublicMiddleware()
	pm.rateLimiter = limiter.NewMemoryLimiter(time.Minute, 1)
	base := time.Unix(1_700_000_000, 0)
	pm.now = func() time.Time { return base }
	handler := pm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	// First request succeeds
	rec1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("GET", "/", nil)
	req1.Header.Set("X-Request-ID", "abc")
	req1.Header.Set("X-API-Timestamp", strconv.FormatInt(base.Unix(), 10))
	req1.Header.Set("X-Forwarded-For", "1.2.3.4")
	if err := handler(rec1, req1); err != nil {
		t.Fatalf("first request failed: %#v", err)
	}

	// Replay with same request ID should be unauthorized
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Request-ID", "abc")
	req2.Header.Set("X-API-Timestamp", strconv.FormatInt(base.Unix(), 10))
	req2.Header.Set("X-Forwarded-For", "1.2.3.4")
	if err := handler(rec2, req2); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for replay, got %#v", err)
	}

	// New request after replay should hit rate limit
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("X-Request-ID", "def")
	req3.Header.Set("X-API-Timestamp", strconv.FormatInt(base.Unix(), 10))
	req3.Header.Set("X-Forwarded-For", "1.2.3.4")
	if err := handler(rec3, req3); err == nil || err.Status != http.StatusTooManyRequests {
		t.Fatalf("expected rate limit error, got %#v", err)
	}
}
