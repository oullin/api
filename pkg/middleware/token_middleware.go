package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"io"
	"log/slog"
	baseHttp "net/http"
	"strings"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cache"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/limiter"
	"github.com/oullin/pkg/portal"
)

const tokenHeader = "X-API-Key"
const usernameHeader = "X-API-Username"
const signatureHeader = "X-API-Signature"
const timestampHeader = "X-API-Timestamp"
const nonceHeader = "X-API-Nonce"
const requestIDHeader = "X-Request-ID"

// Context keys for propagating auth info downstream
// Use unexported custom type to avoid collisions
type contextKey string

const (
	authAccountNameKey contextKey = "auth.account_name"
	requestIdKey       contextKey = "request.id"
)

// TokenCheckMiddleware authenticates signed API requests using account tokens.
// It validates required headers, enforces a timestamp skew window, prevents
// replay attacks via nonce tracking, compares tokens/signatures in constant time,
// and applies a basic failure-based rate limiter per client scope.
//
// Error handling:
// - Rate limiting errors return 429 Too Many Requests
// - Timestamp errors return 401 with specific messages for expired or future timestamps
// - Other authentication errors return 401 with generic messages
type TokenCheckMiddleware struct {
	// ApiKeys provides access to persisted API key records used to resolve
	// account credentials (account name, public key, and secret key).
	ApiKeys *repository.ApiKeys

	// TokenHandler performs encoding/decoding of tokens and signature creation/verification.
	TokenHandler *auth.TokenHandler

	// nonceCache stores recently seen nonce's to prevent replaying the same request
	// within the configured TTL window.
	nonceCache *cache.TTLCache

	// rateLimiter throttles repeated authentication failures per "clientIP|account" scope.
	rateLimiter *limiter.MemoryLimiter

	// clockSkew defines the allowed difference between client and server time when
	// validating the request timestamp.
	clockSkew time.Duration

	// now is an injectable time source for deterministic tests. If nil, time.Now is used.
	now func() time.Time

	// disallowFuture, if true, rejects timestamps greater than the current server time,
	// even if they are within the positive skew window.
	disallowFuture bool

	// nonceTTL is how long a nonce remains invalid after its first use (replay-protection window).
	nonceTTL time.Duration

	// failWindow indicates the sliding time window used to evaluate authentication failures.
	failWindow time.Duration

	// maxFailPerScope is the maximum number of failures allowed within failWindow for a given scope.
	maxFailPerScope int
}

func MakeTokenMiddleware(tokenHandler *auth.TokenHandler, apiKeys *repository.ApiKeys) TokenCheckMiddleware {
	return TokenCheckMiddleware{
		ApiKeys:         apiKeys,
		TokenHandler:    tokenHandler,
		nonceCache:      cache.NewTTLCache(),
		rateLimiter:     limiter.NewMemoryLimiter(1*time.Minute, 10),
		clockSkew:       5 * time.Minute,
		now:             time.Now,
		disallowFuture:  true,
		nonceTTL:        5 * time.Minute,
		failWindow:      1 * time.Minute,
		maxFailPerScope: 10,
	}
}

func (t TokenCheckMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		reqID := strings.TrimSpace(r.Header.Get(requestIDHeader))
		logger := slog.With("request_id", reqID, "path", r.URL.Path, "method", r.Method)

		if reqID == "" || logger == nil {
			return t.getInvalidRequestError()
		}

		if depErr := t.guardDependencies(logger); depErr != nil {
			return depErr
		}

		// Extract and validate required headers
		accountName, publicToken, signature, ts, nonce, hdrErr := t.validateAndGetHeaders(r, logger)
		if hdrErr != nil {
			return hdrErr
		}

		// Validate timestamp within allowed skew using ValidTimestamp helper
		vt := NewValidTimestamp(ts, logger, t.now)
		if tsErr := vt.Validate(t.clockSkew, t.disallowFuture); tsErr != nil {
			return tsErr
		}

		// Read body and compute hash
		bodyHash, bodyErr := t.readBodyHash(r, logger)
		if bodyErr != nil {
			return bodyErr
		}

		// Build canonical request string
		canonical := portal.BuildCanonical(r.Method, r.URL, accountName, publicToken, ts, nonce, bodyHash)

		clientIP := portal.ParseClientIP(r)

		if err := t.shallReject(logger, accountName, publicToken, signature, canonical, nonce, clientIP); err != nil {
			return err
		}

		// Update the request context
		r = t.attachContext(r, accountName, reqID)

		logger.Info("authentication successful")

		return next(w, r)
	}
}

func (t TokenCheckMiddleware) guardDependencies(logger *slog.Logger) *http.ApiError {
	missing := make([]string, 0, 4)

	if t.ApiKeys == nil {
		missing = append(missing, "ApiKeys")
	}

	if t.TokenHandler == nil {
		missing = append(missing, "TokenHandler")
	}

	if t.nonceCache == nil {
		missing = append(missing, "nonceCache")
	}

	if t.rateLimiter == nil {
		missing = append(missing, "rateLimiter")
	}

	if len(missing) > 0 {
		logger.Error("token middleware missing dependencies", "missing", strings.Join(missing, ","))
		return t.getUnauthenticatedError()
	}

	return nil
}

func (t TokenCheckMiddleware) validateAndGetHeaders(r *baseHttp.Request, logger *slog.Logger) (accountName, publicToken, signature, ts, nonce string, apiErr *http.ApiError) {
	accountName = strings.TrimSpace(r.Header.Get(usernameHeader))
	publicToken = strings.TrimSpace(r.Header.Get(tokenHeader))
	signature = strings.TrimSpace(r.Header.Get(signatureHeader))
	ts = strings.TrimSpace(r.Header.Get(timestampHeader))
	nonce = strings.TrimSpace(r.Header.Get(nonceHeader))

	if accountName == "" || publicToken == "" || signature == "" || ts == "" || nonce == "" {
		logger.Warn("missing authentication headers")
		return "", "", "", "", "", t.getInvalidRequestError()
	}

	if err := auth.ValidateTokenFormat(publicToken); err != nil {
		logger.Warn("invalid token format")
		return "", "", "", "", "", t.getInvalidTokenFormatError()
	}

	return accountName, publicToken, signature, ts, nonce, nil
}

func (t TokenCheckMiddleware) readBodyHash(r *baseHttp.Request, logger *slog.Logger) (string, *http.ApiError) {
	if r.Body == nil {
		return portal.Sha256Hex(nil), nil
	}

	b, err := portal.ReadWithSizeLimit(r.Body)
	if err != nil {
		logger.Warn("unable to read body for signing")
		return "", t.getInvalidRequestError()
	}

	// restore for downstream handlers
	r.Body = io.NopCloser(bytes.NewReader(b))

	return portal.Sha256Hex(b), nil
}

func (t TokenCheckMiddleware) attachContext(r *baseHttp.Request, accountName, reqID string) *baseHttp.Request {
	ctx := context.WithValue(r.Context(), authAccountNameKey, accountName)
	ctx = context.WithValue(r.Context(), requestIdKey, reqID)

	return r.WithContext(ctx)
}

func (t TokenCheckMiddleware) shallReject(logger *slog.Logger, accountName, publicToken, signature, canonical, nonce, clientIP string) *http.ApiError {
	limiterKey := clientIP + "|" + strings.ToLower(accountName)

	if t.rateLimiter.TooMany(limiterKey) {
		logger.Warn("too many authentication failures", "ip", clientIP)
		return t.getRateLimitedError()
	}

	var item *database.APIKey
	if item = t.ApiKeys.FindBy(accountName); item == nil {
		t.rateLimiter.Fail(limiterKey)
		logger.Warn("account not found")

		return t.getUnauthenticatedError()
	}

	// Fetch account to understand its keys
	token, err := t.TokenHandler.DecodeTokensFor(
		item.AccountName,
		item.SecretKey,
		item.PublicKey,
	)

	if err != nil {
		t.rateLimiter.Fail(limiterKey)
		logger.Error("failed to decode account keys", "account", item.AccountName, "error", err)

		return t.getUnauthenticatedError()
	}

	// Constant-time compare (fixed-length by hashing) of provided public token vs stored one
	pBytes := []byte(strings.TrimSpace(publicToken))
	eBytes := []byte(strings.TrimSpace(token.PublicKey))
	hP := sha256.Sum256(pBytes)
	hE := sha256.Sum256(eBytes)

	if subtle.ConstantTimeCompare(hP[:], hE[:]) != 1 {
		t.rateLimiter.Fail(limiterKey)
		logger.Warn("public token mismatch", "account", item.AccountName)

		return t.getUnauthenticatedError()
	}

	// Nonce replay protection: atomically check-and-mark (UseOnce)
	if t.nonceCache != nil {
		key := item.AccountName + "|" + nonce

		if t.nonceCache.UseOnce(key, t.nonceTTL) {
			t.rateLimiter.Fail(limiterKey)
			logger.Warn("replay detected: nonce already used", "account", item.AccountName)

			return t.getUnauthenticatedError()
		}
	}

	// Compute local signature over canonical request and compare in constant time (hash to fixed-length first)
	localSignature := auth.CreateSignatureFrom(canonical, token.SecretKey)
	hSig := sha256.Sum256([]byte(strings.TrimSpace(signature)))
	hLocal := sha256.Sum256([]byte(localSignature))

	if subtle.ConstantTimeCompare(hSig[:], hLocal[:]) != 1 {
		t.rateLimiter.Fail(limiterKey)
		logger.Warn("signature mismatch", "account", item.AccountName)

		return t.getUnauthenticatedError()
	}

	return nil
}

func (t TokenCheckMiddleware) getInvalidRequestError() *http.ApiError {
	return &http.ApiError{
		Message: "Invalid authentication headers",
		Status:  baseHttp.StatusUnauthorized,
	}
}

func (t TokenCheckMiddleware) getInvalidTokenFormatError() *http.ApiError {
	return &http.ApiError{
		Message: "Invalid credentials",
		Status:  baseHttp.StatusUnauthorized,
	}
}

func (t TokenCheckMiddleware) getUnauthenticatedError() *http.ApiError {
	return &http.ApiError{
		Message: "Invalid credentials",
		Status:  baseHttp.StatusUnauthorized,
	}
}

func (t TokenCheckMiddleware) getRateLimitedError() *http.ApiError {
	return &http.ApiError{
		Message: "Too many authentication attempts",
		Status:  baseHttp.StatusTooManyRequests,
	}
}

func (t TokenCheckMiddleware) getTimestampTooOldError() *http.ApiError {
	return &http.ApiError{
		Message: "Request timestamp expired",
		Status:  baseHttp.StatusUnauthorized,
	}
}

func (t TokenCheckMiddleware) getTimestampTooNewError() *http.ApiError {
	return &http.ApiError{
		Message: "Request timestamp invalid",
		Status:  baseHttp.StatusUnauthorized,
	}
}
