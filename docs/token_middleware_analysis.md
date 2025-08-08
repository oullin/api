# Token Middleware Analysis and Recommendations

File: pkg/http/middleware/token_middleware.go
Date: 2025-08-08

---

## 1) What it does

The TokenCheckMiddleware enforces a simple HMAC-based request authentication using three custom HTTP headers:

- X-API-Username: The account name registered in the system.
- X-API-Key: The public token (must have pk_ prefix and minimum length).
- X-API-Signature: An HMAC-SHA256 signature computed as HMAC(secret, accountName).

Processing flow:
1. Extracts and trims the three headers; rejects if any is empty.
2. Validates the public token format with auth.ValidateTokenFormat (checks min length and pk_/sk_ prefix).
3. Loads the API key record by account name (case-insensitive) from repository.ApiKeys.
4. Decrypts stored encrypted public/secret tokens using TokenHandler.DecodeTokensFor.
5. Verifies the provided public token equals the decrypted public token.
6. Computes a local signature via auth.CreateSignatureFrom(accountName, secretKey) and compares it to X-API-Signature.
7. On success, logs authentication success and calls the next handler; otherwise returns http.ApiError with Status 401 (generic message).

Notes:
- Secrets are stored encrypted-at-rest (AES-GCM).
- Client-facing errors are generic and do not echo credentials; sensitive details are only in server logs.

---

## 2) What it misses (gaps and limitations)

- No constant-time comparisons:
  - Direct string equality checks for token and signature can leak timing information.

- No replay protection:
  - Signature is static (HMAC(secret, accountName)). If intercepted, it can be replayed indefinitely.

- No request binding:
  - Signature isn’t tied to the specific request (method, path, body, timestamp). MITM can reuse it across endpoints.

- No timestamp and nonce:
  - Lacks X-API-Timestamp and X-API-Nonce to limit replay windows.

- Weak error semantics:
  - Returns 403 for all failures; should use 401 for unauthenticated and reserve 403 for authorization.

- Overly verbose error details:
  - Error messages include masked token/signature and exact account name; still reveals information to a client.

- No audience/scope/role concept:
  - Middleware only authenticates; it doesn’t propagate identity or scopes to downstream authorization.

- No context propagation:
  - Doesn’t set authenticated account/token metadata into request context for later use.

- No rate limiting or lockout:
  - Missing protection against credential stuffing or brute-force on account names.

- No key rotation strategy:
  - There’s no support for multiple active key versions or scheduled rotation.

- No IP/Origin policy:
  - Doesn’t check allowed IP ranges or allowed origins per account.

- Minimal logging / no correlation ID:
  - Logs success but lacks a request ID/correlation ID for tracing and reduced PII in logs.

- No transport security enforcement:
  - Middleware doesn’t enforce HTTPS/mTLS expectations (relies on deployment).

---

## 3) How we can improve it (actionable recommendations)

Quick wins (minimal impact):
- Constant-time compares:
  - Use hmac.Equal or subtle.ConstantTimeCompare for token and signature equality checks.

- Correct status codes:
  - Use 401 Unauthorized for auth failures; keep 403 for later authorization checks.

- Reduce error detail to clients:
  - Return generic messages like "Invalid credentials" without echoing account or tokens.
  - Keep detailed logs server-side with masked values.

- Propagate identity via context:
  - On success, set context values (accountName, apiKeyUUID) for downstream handlers.

- Structured logging and correlation ID:
  - Support/require an X-Request-ID header; log with structured fields and masked secrets.

Security hardening (medium impact):
- Request-bound HMAC signatures:
  - Require clients to sign a canonical string: method + path + query + timestamp + nonce + body-hash.
  - Validate within a short skew window (e.g., ±5 minutes) and reject reused nonces.

- Replay protection:
  - Add headers: X-API-Timestamp (epoch seconds) and X-API-Nonce (random UUID).
  - Track recent nonces per account in a short-lived store (in-memory or Redis) for the timestamp window.

- Input normalization:
  - Canonicalize header casing, path, and query param encoding consistently.

- Rate limiting:
  - Rate limit auth failures per IP/account.

- Key rotation support:
  - Allow multiple active key versions; embed a key ID in the public key (e.g., pk_{kid}_{hash}) or add X-API-Key-ID.

- Tenant policy checks:
  - Optionally enforce allowed IP ranges and origins per account from DB policy.

Stronger assurance options (higher impact):
- mTLS for service-to-service:
  - Use client certs to authenticate server-to-server calls; keep HMAC as a second factor.

- OAuth 2.1 / OIDC for frontend apps:
  - Use Authorization Code with PKCE for browser/mobile; exchange for short-lived access token and refresh token.

- JWTs with short TTL:
  - Issue short-lived JWTs after initial key verification; then rely on JWT for subsequent requests.

- Web Application Firewall (WAF) and TLS enforcement:
  - Enforce HTTPS and add a WAF to mitigate common web attacks.

---

## 4) How it can be hacked (attack scenarios)

- Replay attacks:
  - Since the signature is static per account, an attacker capturing headers once can replay them forever.

- Timing attacks:
  - String equality may leak timing info, helping distinguish valid/invalid tokens/signatures.

- Credential stuffing / enumeration:
  - Uniform error messages but with different latencies can hint whether an account exists.

- MITM / downgrade:
  - If TLS is misconfigured, headers can be intercepted; without timestamp/nonce, replay is trivial.

- Logging leakage:
  - Logs include account names and could include masked tokens; misconfigured logging can leak sensitive info.

- No binding to request details:
  - A captured signature for one endpoint can be replayed on another since signature doesn’t include method/path/body.

- Lack of rate limiting:
  - Attackers can brute-force account names or spam requests without backpressure.

---

## 5) How we can pass less information to the frontend

- Don’t echo credentials:
  - Avoid returning account name, token, or signature in error messages. Use generic client-facing errors.

- Use server-generated correlation IDs:
  - Provide X-Request-ID to frontend for support without revealing auth details.

- Minimize fields in success responses:
  - Only include what the UI needs; avoid returning any API key metadata to the browser.

- Store secrets server-side only:
  - For browser apps, avoid exposing API keys; use session cookies or OAuth tokens instead.

- Differential logging:
  - Keep detailed diagnostics in server logs (masked), not in API responses.

---

## 6) How can we authenticate frontend apps better

For browser-based frontends (SPAs/MPAs):
- Prefer OAuth 2.1 Authorization Code with PKCE + OIDC:
  - Users authenticate with the IdP; the SPA exchanges the code for short-lived access tokens and refresh tokens via a BFF (Backend-for-Frontend) to avoid exposing refresh tokens to JS.

- Session cookies with SameSite=strict, HttpOnly, Secure:
  - Use server-managed sessions; issue short-lived session cookies and rotate session IDs frequently.

- Token lifetimes and rotation:
  - Access tokens 5–15 minutes; refresh tokens 7–30 days with rotation and revocation.

- BFF pattern:
  - The frontend talks to your BFF; the BFF calls the API with service credentials, keeping secrets off the browser.

For native apps or trusted server-to-server clients:
- mTLS:
  - Bind clients via mutual TLS certificates.

- Signed requests (HMAC) with request binding:
  - Include method, path, timestamp, nonce, and payload hash; enforce a skew window and nonce cache.

- Device-bound credentials:
  - Use secure enclave/Keychain/TPM to store tokens and bind them to devices.

---

## 7) Suggested phased plan (Checklist)

- [x] Phase 1 (Low risk, immediate)
  - [x] A1. Switch to constant-time comparisons for signature and public token.
  - [x] A2. Return 401 for authentication failures; generic error messages to clients.
  - [x] A3. Add structured logging with X-Request-ID; mask all sensitive values.
  - [x] A4. Put authenticated account into request context.

- [ ] Phase 2 (Security hardening)
  - [ ] B1. Add X-API-Timestamp and X-API-Nonce headers, validate clock skew.
  - [ ] B2. Introduce nonce replay cache (in-memory or Redis) keyed by account+nonce within the time window.
  - [ ] B3. Define canonical request string and require clients to sign it with HMAC(secret, canonical_request).
  - [ ] B4. Add rate limiting on failed auth per IP/account.

- [ ] Phase 3 (Operational maturity)
  - [ ] C1. Implement key rotation with key IDs; allow overlapping validity windows.
  - [ ] C2. Optional IP allowlist/origin policy per account.
  - [ ] C3. mTLS for backend integrations where applicable.

- [ ] Phase 4 (Frontend modernization)
  - [ ] D1. Adopt OAuth 2.1 Authorization Code with PKCE for browser/mobile apps.
  - [ ] D2. Introduce a BFF to keep tokens and secrets off the browser.

---

## 8) Example canonical signature spec (for future adoption)

Headers required:
- X-API-Username
- X-API-Key
- X-API-Timestamp (epoch seconds)
- X-API-Nonce (UUID v4)
- X-API-Signature

Canonical request (string to sign):

METHOD + "\n" +
PATH + "\n" +
SORTED_QUERY_STRING + "\n" +
X-API-Username + "\n" +
X-API-Key + "\n" +
X-API-Timestamp + "\n" +
X-API-Nonce + "\n" +
SHA256_HEX(BODY)

Signature:
- signature = hex(HMAC-SHA256(secretKey, canonical_request))

Validation rules:
- Accept if |now - timestamp| <= 300s, nonce unused within window, and constant-time comparison passes.

---

## 9) Logging guidelines

- Never log full tokens or signatures. Use auth.SafeDisplay or stricter masking.
- Include: request_id, account_name (normalized), result (success/failure), reason codes, client_ip (if safe), user_agent (optional), path, method, and timing.
- Store detailed diagnostics server-side only; respond to clients with generic messages.

---

## 10) References

- RFC 6749 (OAuth 2.0), OAuth 2.1 draft best practices.
- RFC 7636 (PKCE).
- NIST SP 800-63B (Digital Identity Guidelines).
- Timing attack mitigations (use constant-time comparisons).
