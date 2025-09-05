package middleware

import (
	"context"
	"encoding/hex"
	"fmt"
	baseHttp "net/http"
	"strings"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/repoentity"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cache"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/limiter"
	"github.com/oullin/pkg/middleware/mwguards"
	"github.com/oullin/pkg/portal"
)

const tokenHeader = "X-API-Key"
const usernameHeader = "X-API-Username"
const signatureHeader = "X-API-Signature"
const timestampHeader = "X-API-Timestamp"
const nonceHeader = "X-API-Nonce"
const requestIDHeader = "X-Request-ID"
const intendedOrigin = "X-API-Intended-Origin"

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

	// nonceCache stores recently seen nonce to prevent replaying the same request
	// within the configured TTL window.
	nonceCache *cache.TTLCache

	// rateLimiter throttles repeated authentication failures per "clientIP|account" scope.
	rateLimiter *limiter.MemoryLimiter

	// clockSkew defines the allowed difference between client and server time when
	// validating the request timestamp.
	clockSkew time.Duration

	// Now is an injectable time source for deterministic tests. If nil, time.Now is used.
	now func() time.Time

	// disallowFuture, if true, rejects timestamps greater than the current server time,
	// even if they are within the positive skew window.
	disallowFuture bool

	// nonceTTL is how long nonce remains invalid after its first use (replay-protection window).
	nonceTTL time.Duration

	// failWindow indicates the sliding time window used to evaluate authentication failures.
	failWindow time.Duration

	// maxFailPerScope is the maximum number of failures allowed within the failWindow for a given scope.
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

		if reqID == "" {
			return mwguards.InvalidRequestError(fmt.Sprintf("Invalid request ID for URL [%s].", r.URL.Path), "")
		}

		if err := t.GuardDependencies(); err != nil {
			return err
		}

		headers, err := t.ValidateAndGetHeaders(r, reqID)
		if err != nil {
			return err
		}

		// Validate timestamp within allowed skew using ValidTimestamp helper
		vt := NewValidTimestamp(headers.Timestamp, t.now)
		if tsErr := vt.Validate(t.clockSkew, t.disallowFuture); tsErr != nil {
			return tsErr
		}

		var apiKey *database.APIKey
		if apiKey, err = t.HasInvalidFormat(headers); err != nil {
			return err
		}

		if err = t.HasInvalidSignature(headers, apiKey); err != nil {
			return err
		}

		r = t.AttachContext(r, headers)

		return next(w, r)
	}
}

func (t TokenCheckMiddleware) GuardDependencies() *http.ApiError {
	missing := make([]string, 0, 4)

	if t.ApiKeys == nil {
		missing = append(missing, "KeysRepository")
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
		return mwguards.UnauthenticatedError(
			"token middleware missing dependencies",
			"token middleware missing dependencies: "+strings.Join(missing, ",")+".",
			map[string]any{
				"missing": missing,
			},
		)
	}

	return nil
}

func (t TokenCheckMiddleware) ValidateAndGetHeaders(r *baseHttp.Request, requestId string) (AuthTokenHeaders, *http.ApiError) {
	accountName := strings.TrimSpace(r.Header.Get(usernameHeader))
	signature := strings.TrimSpace(r.Header.Get(signatureHeader))
	publicToken := strings.TrimSpace(r.Header.Get(tokenHeader))
	ts := strings.TrimSpace(r.Header.Get(timestampHeader))
	nonce := strings.TrimSpace(r.Header.Get(nonceHeader))
	ip := portal.ParseClientIP(r)
	intendedOriginURL := strings.TrimSpace(r.Header.Get(intendedOrigin))

	if accountName == "" || publicToken == "" || signature == "" || ts == "" || nonce == "" || ip == "" || intendedOriginURL == "" {
		return AuthTokenHeaders{}, mwguards.InvalidRequestError(
			"Invalid authentication headers / or missing headers",
			"",
		)
	}

	if err := auth.ValidateTokenFormat(publicToken); err != nil {
		return AuthTokenHeaders{}, mwguards.InvalidTokenFormatError(err.Error(), "", map[string]any{})
	}

	return AuthTokenHeaders{
		AccountName:       accountName,
		PublicKey:         publicToken,
		Signature:         signature,
		Timestamp:         ts,
		Nonce:             nonce,
		ClientIP:          ip,
		RequestID:         requestId,
		IntendedOriginURL: intendedOriginURL,
	}, nil
}

func (t TokenCheckMiddleware) AttachContext(r *baseHttp.Request, headers AuthTokenHeaders) *baseHttp.Request {
	ctx := context.WithValue(r.Context(), authAccountNameKey, headers.AccountName)
	ctx = context.WithValue(r.Context(), requestIdKey, headers.RequestID)

	return r.WithContext(ctx)
}

func (t TokenCheckMiddleware) HasInvalidFormat(headers AuthTokenHeaders) (*database.APIKey, *http.ApiError) {
	limiterKey := headers.ClientIP + "|" + strings.ToLower(headers.AccountName)

	if t.rateLimiter.TooMany(limiterKey) {
		return nil, mwguards.RateLimitedError(
			"Too many authentication attempts",
			"Too many authentication attempts for key: "+limiterKey,
		)
	}

	guard := mwguards.NewMWTokenGuard(t.ApiKeys, t.TokenHandler)

	rejectsRequest := mwguards.MWTokenGuardData{
		Username:  headers.AccountName,
		PublicKey: headers.PublicKey,
	}

	if guard.Rejects(rejectsRequest) {
		t.rateLimiter.Fail(headers.AccountName)

		return nil, mwguards.UnauthenticatedError(
			"Invalid public token",
			guard.Error.Error(),
			rejectsRequest.ToMap(),
		)
	}

	if t.nonceCache != nil {
		key := strings.ToLower(headers.AccountName) + "|" + headers.Nonce

		if t.nonceCache.UseOnce(key, t.nonceTTL) {
			t.rateLimiter.Fail(limiterKey)

			return nil, mwguards.UnauthenticatedError(
				"Invalid nonce",
				"Invalid nonce using key: "+key,
				map[string]any{"key": key, "limiter_key": limiterKey},
			)
		}
	}

	return guard.ApiKey, nil
}

func (t TokenCheckMiddleware) HasInvalidSignature(headers AuthTokenHeaders, apiKey *database.APIKey) *http.ApiError {
	var err error
	var byteSignature []byte

	if byteSignature, err = hex.DecodeString(headers.Signature); err != nil {
		return mwguards.NotFound("error decoding signature string", "")
	}

	entity := repoentity.FindSignatureFrom{
		Key:        apiKey,
		Signature:  byteSignature,
		Origin:     headers.IntendedOriginURL,
		ServerTime: time.Now(),
	}

	signature := t.ApiKeys.FindSignatureFrom(entity)

	if signature == nil {
		return mwguards.NotFound("signature not found", "")
	}

	if err = t.ApiKeys.IncreaseSignatureTries(signature.UUID, signature.CurrentTries+1); err != nil {
		return mwguards.InvalidRequestError("could not increase signature tries", err.Error())
	}

	return nil
}
