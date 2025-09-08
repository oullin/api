package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	pkgHttp "github.com/oullin/pkg/http"
	"github.com/oullin/pkg/limiter"
)

func TestPublicMiddleware_InvalidHeaders(t *testing.T) {
	pm := MakePublicMiddleware([]byte("test-secret"))
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
	pm := MakePublicMiddleware([]byte("test-secret"))
	base := time.Unix(1_700_000_000, 0)
	pm.now = func() time.Time { return base }
	handler := pm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-1")
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	old := base.Add(-10 * time.Minute).Unix()
	req.Header.Set("X-API-Timestamp", strconv.FormatInt(old, 10))
	req.Header.Set("X-API-Signature", sign([]byte("test-secret"), "req-1", strconv.FormatInt(old, 10), "1.2.3.4"))
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for old timestamp, got %#v", err)
	}
}

func TestPublicMiddleware_RateLimitAndReplay(t *testing.T) {
	pm := MakePublicMiddleware([]byte("test-secret"))
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
	req1.Header.Set("X-API-Signature", sign([]byte("test-secret"), "abc", strconv.FormatInt(base.Unix(), 10), "1.2.3.4"))
	if err := handler(rec1, req1); err != nil {
		t.Fatalf("first request failed: %#v", err)
	}

	// Replay with same request ID should be unauthorized
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("X-Request-ID", "abc")
	req2.Header.Set("X-API-Timestamp", strconv.FormatInt(base.Unix(), 10))
	req2.Header.Set("X-Forwarded-For", "1.2.3.4")
	req2.Header.Set("X-API-Signature", sign([]byte("test-secret"), "abc", strconv.FormatInt(base.Unix(), 10), "1.2.3.4"))
	if err := handler(rec2, req2); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for replay, got %#v", err)
	}

	// New request after replay should hit rate limit
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("X-Request-ID", "def")
	req3.Header.Set("X-API-Timestamp", strconv.FormatInt(base.Unix(), 10))
	req3.Header.Set("X-Forwarded-For", "1.2.3.4")
	req3.Header.Set("X-API-Signature", sign([]byte("test-secret"), "def", strconv.FormatInt(base.Unix(), 10), "1.2.3.4"))
	if err := handler(rec3, req3); err == nil || err.Status != http.StatusTooManyRequests {
		t.Fatalf("expected rate limit error, got %#v", err)
	}
}

func TestPublicMiddleware_InvalidSignature(t *testing.T) {
	pm := MakePublicMiddleware([]byte("test-secret"))
	base := time.Unix(1_700_000_000, 0)
	pm.now = func() time.Time { return base }
	handler := pm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-1")
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	ts := strconv.FormatInt(base.Unix(), 10)
	req.Header.Set("X-API-Timestamp", ts)
	// incorrect signature
	req.Header.Set("X-API-Signature", sign([]byte("other"), "req-1", ts, "1.2.3.4"))
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for bad signature, got %#v", err)
	}
}

func sign(secret []byte, reqID, ts, ip string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(reqID + "|" + ts + "|" + ip))
	return hex.EncodeToString(mac.Sum(nil))
}
