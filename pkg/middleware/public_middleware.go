package middleware

import (
	"fmt"
	baseHttp "net/http"
	"strings"
	"time"

	"github.com/oullin/pkg/cache"
	"github.com/oullin/pkg/endpoint"
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
	now            func() time.Time
	allowedIP      string
	isProduction   bool
}

func MakePublicMiddleware(allowedIP string, isProduction bool) PublicMiddleware {
	return PublicMiddleware{
		clockSkew:      5 * time.Minute,
		disallowFuture: true,
		requestTTL:     5 * time.Minute,
		rateLimiter:    limiter.NewMemoryLimiter(1*time.Minute, 10),
		requestCache:   cache.NewTTLCache(),
		now:            time.Now,
		allowedIP:      strings.TrimSpace(allowedIP),
		isProduction:   isProduction,
	}
}

func (p PublicMiddleware) Handle(next endpoint.ApiHandler) endpoint.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *endpoint.ApiError {
		if err := p.GuardDependencies(); err != nil {
			return err
		}

		uri := portal.GenerateURL(r)
		ts := strings.TrimSpace(r.Header.Get(portal.TimestampHeader))
		reqID := strings.TrimSpace(r.Header.Get(portal.RequestIDHeader))

		if reqID == "" || ts == "" {
			return mwguards.InvalidRequestError("Invalid authentication headers", "")
		}

		limiterKey := strings.Join([]string{uri, reqID, ts}, "|")

		if p.rateLimiter.TooMany(limiterKey) {
			return mwguards.RateLimitedError("Too many requests", "Too many requests for key: "+limiterKey)
		}

		vt := mwguards.NewValidTimestamp(ts, p.now)
		if err := vt.Validate(p.clockSkew, p.disallowFuture); err != nil {
			p.rateLimiter.Fail(limiterKey)

			return err
		}

		key := strings.Join([]string{limiterKey, reqID}, "|")
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

func (p PublicMiddleware) GuardDependencies() *endpoint.ApiError {
	missing := []string{}

	if p.requestCache == nil {
		missing = append(missing, "requestCache")
	}

	if p.rateLimiter == nil {
		missing = append(missing, "rateLimiter")
	}

	if len(missing) > 0 {
		err := fmt.Errorf("public middleware missing dependencies: %s", strings.Join(missing, ","))
		return endpoint.LogInternalError("public middleware missing dependencies", err)
	}

	return nil
}
