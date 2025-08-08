# Token middleware analysis (v2)

Date: 2025-08-08
Scope: pkg/middleware/token_middleware.go and related helpers (valid_timestamp.go, pkg/portal/support.go)

---

## 1) What the token middleware does today

The TokenCheckMiddleware authenticates signed API requests using per‑account API keys. It enforces a signed-request protocol that binds each request to:
- HTTP method and path
- Sorted query string
- Account username and public token
- Unix timestamp and nonce
- Body content hash

Main steps:
1) Required headers
   - X-Request-ID (required for logging/tracing)
   - X-API-Username (account identifier)
   - X-API-Key (public token)
   - X-API-Signature (HMAC over canonical request)
   - X-API-Timestamp (Unix seconds)
   - X-API-Nonce (unique per request)

2) Dependency guard
   - Ensures ApiKeys repo, TokenHandler, nonce cache, and rate limiter exist. If missing, fails with 401.

3) Header validation
   - Rejects if any required header is missing (401: "Invalid authentication headers").
   - Validates public token format via auth.ValidateTokenFormat (e.g., prefix/length conventions); on failure returns generic 401 ("Invalid credentials").

4) Timestamp validation
   - Uses ValidTimestamp.Validate with a configurable skew (default 5 minutes) and an optional disallowFuture flag (default false).
   - Rejects if timestamp older than now - skew or (if disallowFuture) newer than now; otherwise allows within [now - skew, now + skew].

5) Body hashing
   - Reads the request body to compute SHA-256 hex via portal.Sha256Hex, then restores the body so downstream handlers can read it again.

6) Canonical request construction
   - portal.BuildCanonical(method, url, username, public, ts, nonce, bodyHash)
   - Includes: uppercased method, escaped path (default "/"), sorted & url-escaped query string, username, public token, timestamp, nonce, body hash, joined by newlines.

7) Client IP parsing
   - portal.ParseClientIP prefers X-Forwarded-For first IP, otherwise uses RemoteAddr host.

8) Rejection checks (shallReject)
   - Failure-based rate limiting per scope clientIP|account (MemoryLimiter with 1-minute window and max 10 fails):
     - If TooMany, reject early.
   - Account lookup via repository.ApiKeys.FindBy(username); if not found, mark failure and reject.
   - Key decoding via TokenHandler.DecodeTokensFor(account, encSecret, encPublic); on error, mark failure and reject.
   - Constant-time compare (crypto/subtle.ConstantTimeCompare) of provided public token vs decoded public.
   - Nonce replay protection with TTL cache: if nonce for account already used within TTL, reject; otherwise mark it as used after successful signature.
   - HMAC signature verification: localSignature = auth.CreateSignatureFrom(canonical, token.SecretKey); constant-time compare vs provided signature.

9) Context propagation
   - Attaches auth.account_name and request.id to context for downstream handlers.

10) Logging and errors
   - Uses structured slog logger with request_id, method, path.
   - Logs warnings for missing headers, invalid format, too many failures, account not found, replay detected, signature mismatch.
   - Errors are generic to clients (HTTP 401 with neutral messages) to avoid leaking details.

Defaults (from MakeTokenMiddleware):
- clockSkew: 5m; disallowFuture: false
- nonceTTL: 5m; nonceCache: in-memory TTL
- rateLimiter: in-memory with 1m window, 10 failures threshold
- now: time.Now (injectable for tests)

---

## 2) What “Version 1” achieved

From the implementation and tests, v1 delivered the following:
- Request binding and signing
  - Canonical request includes method, path, sorted query, timestamp, nonce, and body hash.
  - HMAC signature over canonical string using per-account secret key.
- Input hardening
  - Strict header presence checks.
  - Token format validation.
  - Constant-time comparisons for public token and signature to prevent timing leaks.
- Replay and freshness controls
  - Timestamp skew window enforcement with configurable policy.
  - Nonce replay cache with TTL per account.
- Operational safeguards
  - Failure-based rate limiting per clientIP|account scope.
  - Structured logging with request correlation id (X-Request-ID) and neutral client-facing errors.
- DX and correctness
  - Body is re-usable after hashing.
  - Context carries account name and request id for downstream.
  - Thorough unit/integration tests (including canonicalization and DB-backed key lookup/decoding).

These map to the earlier docs’ Phases 1–2 checkboxes: constant-time compare, generic 401s, structured logging with request-id, context propagation, timestamp & nonce controls, canonical signing, failure rate limiting.

---

## 3) Gaps and risks

- In-memory state
  - Nonce cache and rate limiter are in-memory; not horizontally scalable. Replays could work across nodes, and failure limits won’t coordinate cluster-wide.
- Future timestamp policy
  - disallowFuture is false by default, allowing future timestamps within skew; opens small window for limited replay across clocks.
- Key rotation and key identification
  - Canonical/signature doesn’t include a key identifier (kid). Rotations require coordination and lookup; currently only username+public token are provided.
- Observability and audit
  - Logs are present but no explicit audit event for successful/failed auth with stable event schema or metrics emission.
- Error handling surfaces
  - All errors map to 401; may want 429 for rate-limited scopes and distinct codes for clock skew vs. formatting for better client remediation (while still generic).
- DoS controls
  - Rate limiter is failure-triggered only; no general request token bucket per account/IP for auth endpoints.
- Request size / body hashing
  - Large bodies are fully read into memory to hash; could be abused. No explicit max size before rejection.

---

## 4) Phase 3: Recommended next steps

1) Distributed nonce and rate limiting
   - Replace in-memory TTL nonce cache with Redis (SETNX with TTL) keyed by account|nonce.
   - Use Redis or a shared backend for failure-based limiter; emit 429 Too Many Requests when threshold exceeded.

2) Tighten time policy and drift support
   - Set disallowFuture = true by default; document policy for clients.
   - Reduce skew to 2 minutes (configurable), and expose a time sync endpoint returning server time (and maybe signed) to help clients adjust.

3) Key management and rotation
   - Introduce key IDs (kid) and include it in headers and canonical string.
   - Support multiple active keys per account with activation/expiration and server-side rotation policy.
   - Add a deprecation window and telemetry to observe usage of old keys before revocation.

4) Stronger request binding and algorithm agility
   - Version the signing scheme (alg/version) in headers; allow upgrading to stronger algorithms if needed.
   - Ensure canonicalization is frozen per version and formally documented; add conformance tests for edge cases (empty query, repeated params, unicode path).

5) Resource and abuse protections
   - Enforce a maximum body size for requests that must be signed (e.g., 2–5MB) before hashing; reject larger with 413.
   - Add general-purpose rate limiting (token bucket) per IP/account for auth attempts and overall requests.

6) Observability and audit
   - Emit structured audit events (success/failure) with fields: request_id, account, client_ip (anonymized), reason, kid, alg, and skew/nonce metrics.
   - Add metrics (Prometheus): auth_success_total, auth_failure_total by reason, replay_detected_total, skew_violations_total, limiter_block_total.

7) Security posture improvements
   - Consider mTLS for server-to-server clients; retain HMAC as app-level assurance.
   - Add IP allow/deny lists for sensitive routes.
   - Ensure all secrets are stored encrypted at rest and rotate encryption keys for stored API keys.

8) Developer ergonomics and docs
   - Publish exact canonicalization rules and client libraries/examples for multiple languages.
   - Provide a sandbox endpoint to verify signatures and surface detailed diagnostics to authenticated developers.

9) Backward compatibility and rollout
   - Introduce phase-3 features behind config flags.
   - Add dual-signing period when introducing kid/versioned algorithms so clients can switch gradually.

---

## 5) Concrete implementation tasks

- Replace nonceCache with interface and add RedisTTLCache implementation; wire by config. Use key "acct|nonce" with PX TTL and SET NX.
- Replace limiter.MemoryLimiter with a pluggable Limiter interface and add Redis-based sliding window or token bucket implementation. Return 429 on scope saturation.
- Default disallowFuture = true; make skew configurable via env. Add /time endpoint that returns server unix time.
- Extend headers:
  - X-API-Key-ID (kid)
  - X-API-Alg (e.g., hmac-sha256;v=1)
  Include these in canonical string and signature verification path. Update tests.
- Add MaxBodyBytes wrapper before hashing; reject with 413 if exceeded; document limits.
- Emit metrics and structured audit logs; create middleware counters and reason tags.
- Documentation updates and client examples for canonicalization and signing including new headers.

---

## 6) What remains unchanged (for clarity)

- The overall request flow and canonicalization approach remains; we are enriching it with distributed state, policy tightening, and key management features.
- Error messages remain generic to clients; only status codes and headers may change for rate-limiting (429) and possibly clock skew guidance headers.

---

## 7) Testing strategy for Phase 3

- Unit tests
  - Canonicalization invariants, header parsing with kid/alg, disallowFuture behavior, max body enforcement.
- Integration tests
  - Redis-backed nonce and limiter behavior across multiple requests and simulated nodes.
  - Key rotation: old/new key acceptance within window, rejection after revocation.
- Load tests
  - Validate limiter correctness under concurrency and replay attempts.

