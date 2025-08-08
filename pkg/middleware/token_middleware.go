package middleware

import (
	"bytes"
	"context"
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
		nonceTTL:        5 * time.Minute,
		failWindow:      1 * time.Minute,
		maxFailPerScope: 10,
	}
}

func (t TokenCheckMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		reqID := strings.TrimSpace(r.Header.Get(requestIDHeader))
		if reqID == "" {
			return t.getInvalidRequestError()
		}

		logger := slog.With("request_id", reqID, "path", r.URL.Path, "method", r.Method)

		// Extract and validate required headers
		accountName, publicToken, signature, ts, nonce, hdrErr := t.validateAndGetHeaders(r, logger)
		if hdrErr != nil {
			return hdrErr
		}

		// Validate timestamp within allowed skew
		if tsErr := t.validateTimestamp(ts, logger); tsErr != nil {
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

		if t.shallReject(logger, accountName, publicToken, signature, canonical, nonce, clientIP) {
			return t.getUnauthenticatedError()
		}

		// Update the request context
		r = t.attachAuthContext(r, accountName, reqID)

		logger.Info("authentication successful")

		return next(w, r)
	}
}

// validateAndGetHeaders extracts and validates required auth headers, logging and returning
// appropriate ApiError on failure.
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

// validateTimestamp ensures the provided timestamp is numeric and within skew.
func (t TokenCheckMiddleware) validateTimestamp(ts string, logger *slog.Logger) *http.ApiError {
	if ts == "" {
		logger.Warn("missing timestamp")
		return t.getInvalidRequestError()
	}

	var epoch int64
	for _, ch := range ts {
		if ch < '0' || ch > '9' {
			logger.Warn("invalid timestamp format")
			return t.getInvalidRequestError()
		}

		epoch = epoch*10 + int64(ch-'0')
	}

	now := time.Now().Unix()
	if epoch < now-int64(t.clockSkew.Seconds()) || epoch > now+int64(t.clockSkew.Seconds()) {
		logger.Warn("timestamp outside allowed window")
		return t.getUnauthenticatedError()
	}

	return nil
}

// readBodyHash reads and restores the request body and returns its SHA256 hex.
func (t TokenCheckMiddleware) readBodyHash(r *baseHttp.Request, logger *slog.Logger) (string, *http.ApiError) {
	if r.Body == nil {
		return portal.Sha256Hex(nil), nil
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Warn("unable to read body for signing")
		return "", t.getInvalidRequestError()
	}

	// restore for downstream handlers
	r.Body = io.NopCloser(bytes.NewReader(b))

	return portal.Sha256Hex(b), nil
}

// attachAuthContext adds auth/account data and request id to the request context.
func (t TokenCheckMiddleware) attachAuthContext(r *baseHttp.Request, accountName, reqID string) *baseHttp.Request {
	ctx := context.WithValue(r.Context(), authAccountNameKey, accountName)
	ctx = context.WithValue(r.Context(), requestIdKey, reqID)
	return r.WithContext(ctx)
}

func (t TokenCheckMiddleware) shallReject(logger *slog.Logger, accountName, publicToken, signature, canonical, nonce, clientIP string) bool {
	limiterKey := clientIP + "|" + strings.ToLower(accountName)

	if t.rateLimiter != nil && t.rateLimiter.TooMany(limiterKey) {
		logger.Warn("too many authentication failures", "ip", clientIP)
		return true
	}

	var item *database.APIKey
	if item = t.ApiKeys.FindBy(accountName); item == nil {
		if t.rateLimiter != nil {
			t.rateLimiter.Fail(limiterKey)
		}

		logger.Warn("account not found")
		return true
	}

	token, err := t.TokenHandler.DecodeTokensFor(
		item.AccountName,
		item.SecretKey,
		item.PublicKey,
	)

	if err != nil {
		if t.rateLimiter != nil {
			t.rateLimiter.Fail(limiterKey)
		}
		logger.Error("failed to decode account keys", "account", item.AccountName, "error", err)
		return true
	}

	// Constant-time compare of provided public token vs stored one
	provided := []byte(strings.TrimSpace(publicToken))
	expected := []byte(strings.TrimSpace(token.PublicKey))
	if subtle.ConstantTimeCompare(provided, expected) != 1 {
		if t.rateLimiter != nil {
			t.rateLimiter.Fail(limiterKey)
		}
		logger.Warn("public token mismatch", "account", item.AccountName)
		return true
	}

	// Nonce replay protection: check if already used for this account
	if t.nonceCache != nil {
		key := item.AccountName + "|" + nonce
		if t.nonceCache.Used(key) {
			if t.rateLimiter != nil {
				t.rateLimiter.Fail(limiterKey)
			}

			logger.Warn("replay detected: nonce already used", "account", item.AccountName)
			return true
		}
	}

	// Compute local signature over canonical request and compare in constant time
	localSignature := auth.CreateSignatureFrom(canonical, token.SecretKey)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(localSignature)) != 1 {
		if t.rateLimiter != nil {
			t.rateLimiter.Fail(limiterKey)
		}
		logger.Warn("signature mismatch", "account", item.AccountName)
		return true
	}

	// Mark nonce as used within the TTL
	if t.nonceCache != nil {
		key := item.AccountName + "|" + nonce
		t.nonceCache.Mark(key, t.nonceTTL)
	}

	return false
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
