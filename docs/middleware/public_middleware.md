# Public middleware

The `PublicMiddleware` protects openly accessible endpoints with
lightweight in-memory defenses. It is defined in
`pkg/middleware/public_middleware.go` and provides:

- **Rate limiting** – `limiter.MemoryLimiter` caps requests per client
  IP within a sliding window.
- **Timestamp validation** – `ValidTimestamp` ensures the
  `X-API-Timestamp` header is within an allowed skew (5 minutes by
default).
- **Replay protection** – a `cache.TTLCache` tracks used request IDs and
  rejects duplicates using a composite key of
  `limiterKey|requestID|ip` (rate limiter key, request ID and client IP).
- **Dependency checks** – missing caches or limiters are logged and cause
  a generic 500 Internal Server Error.

### Required headers

- `X-Request-ID`
- `X-API-Timestamp`

Requests lacking these headers, using an unparsable client IP, or failing
validation are rejected with an authentication error. Valid requests pass
through to the next handler.
