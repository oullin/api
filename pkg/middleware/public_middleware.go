package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	baseHttp "net/http"
	"strings"
	"time"

	"github.com/oullin/pkg/cache"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/limiter"
	"github.com/oullin/pkg/middleware/mwguards"
	"github.com/oullin/pkg/portal"
)

// PublicMiddleware provides basic protections for public endpoints.
// It enforces a timestamp check to prevent replay attacks and applies
// a simple in-memory rate limiter keyed by client IP. Reuse of a
// request ID within a TTL window is rejected via TTLCache.
type PublicMiddleware struct {
	clockSkew      time.Duration
	disallowFuture bool
	requestTTL     time.Duration
	rateLimiter    *limiter.MemoryLimiter
	requestCache   *cache.TTLCache
	signingSecret  []byte
	now            func() time.Time
}

// MakePublicMiddleware constructs a PublicMiddleware with sane defaults.
// A non-nil signing secret is required to validate request signatures.
func MakePublicMiddleware(secret []byte) PublicMiddleware {
	return PublicMiddleware{
		clockSkew:      5 * time.Minute,
		disallowFuture: true,
		requestTTL:     5 * time.Minute,
		rateLimiter:    limiter.NewMemoryLimiter(1*time.Minute, 10),
		requestCache:   cache.NewTTLCache(),
		signingSecret:  secret,
		now:            time.Now,
	}
}

func (p PublicMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		if err := p.guardDependencies(); err != nil {
			return err
		}

		reqID := strings.TrimSpace(r.Header.Get(portal.RequestIDHeader))
		ts := strings.TrimSpace(r.Header.Get(portal.TimestampHeader))
		if reqID == "" || ts == "" {
			return mwguards.InvalidRequestError("Invalid authentication headers", "")
		}

		ip := portal.ParseClientIP(r)
		if ip == "" {
			return mwguards.InvalidRequestError("Invalid client IP", "")
		}

		limiterKey := ip
		if p.rateLimiter.TooMany(limiterKey) {
			return mwguards.RateLimitedError("Too many requests", "Too many requests for key: "+limiterKey)
		}

		vt := NewValidTimestamp(ts, p.now)
		if err := vt.Validate(p.clockSkew, p.disallowFuture); err != nil {
			return err
		}

		sig := strings.TrimSpace(r.Header.Get(portal.SignatureHeader))
		if sig == "" {
			return mwguards.InvalidRequestError("Invalid authentication headers", "")
		}

		payload := reqID + "|" + ts + "|" + ip
		mac := hmac.New(sha256.New, p.signingSecret)
		mac.Write([]byte(payload))
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(strings.ToLower(sig)), []byte(expected)) {
			p.rateLimiter.Fail(limiterKey)
			return mwguards.UnauthenticatedError(
				"Invalid signature",
				"signature mismatch",
				map[string]any{"limiter_key": limiterKey},
			)
		}

		key := limiterKey + "|" + reqID + ip
		if p.requestCache.UseOnce(key, p.requestTTL) {
			p.rateLimiter.Fail(limiterKey)
			return mwguards.UnauthenticatedError(
				"Invalid request id",
				"duplicate request id: "+key,
				map[string]any{"key": key, "limiter_key": limiterKey},
			)
		}

		return next(w, r)
	}
}

func (p PublicMiddleware) guardDependencies() *http.ApiError {
	missing := []string{}
	if p.requestCache == nil {
		missing = append(missing, "requestCache")
	}
	if p.rateLimiter == nil {
		missing = append(missing, "rateLimiter")
	}
	if len(p.signingSecret) == 0 {
		missing = append(missing, "signingSecret")
	}
	if len(missing) > 0 {
		return mwguards.UnauthenticatedError(
			"public middleware missing dependencies",
			"public middleware missing dependencies: "+strings.Join(missing, ","),
			map[string]any{"missing": missing},
		)
	}
	return nil
}
