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

type TokenCheckMiddleware struct {
	maxFailPerScope int
	disallowFuture  bool
	nonceTTL        time.Duration
	failWindow      time.Duration
	clockSkew       time.Duration
	nonceCache      *cache.TTLCache
	now             func() time.Time
	TokenHandler    *auth.TokenHandler
	ApiKeys         *repository.ApiKeys
	rateLimiter     *limiter.MemoryLimiter
}

func MakeTokenMiddleware(tokenHandler *auth.TokenHandler, apiKeys *repository.ApiKeys) TokenCheckMiddleware {
	return TokenCheckMiddleware{
		maxFailPerScope: 10,
		disallowFuture:  true,
		ApiKeys:         apiKeys,
		now:             time.Now,
		TokenHandler:    tokenHandler,
		clockSkew:       5 * time.Minute,
		failWindow:      1 * time.Minute,
		nonceTTL:        5 * time.Minute,
		nonceCache:      cache.NewTTLCache(),
		rateLimiter:     limiter.NewMemoryLimiter(1*time.Minute, 10),
	}
}

func (t TokenCheckMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		reqID := strings.TrimSpace(r.Header.Get(portal.RequestIDHeader))

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
	intendedOriginURL := strings.TrimSpace(r.Header.Get(portal.IntendedOrigin))
	accountName := strings.TrimSpace(r.Header.Get(portal.UsernameHeader))
	signature := strings.TrimSpace(r.Header.Get(portal.SignatureHeader))
	publicToken := strings.TrimSpace(r.Header.Get(portal.TokenHeader))
	ts := strings.TrimSpace(r.Header.Get(portal.TimestampHeader))
	nonce := strings.TrimSpace(r.Header.Get(portal.NonceHeader))
	ip := portal.ParseClientIP(r)

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
		Timestamp:         ts,
		ClientIP:          ip,
		Nonce:             nonce,
		Signature:         signature,
		RequestID:         requestId,
		AccountName:       accountName,
		PublicKey:         publicToken,
		IntendedOriginURL: intendedOriginURL,
	}, nil
}

func (t TokenCheckMiddleware) AttachContext(r *baseHttp.Request, headers AuthTokenHeaders) *baseHttp.Request {
	ctx := context.WithValue(r.Context(), portal.AuthAccountNameKey, headers.AccountName)
	ctx = context.WithValue(r.Context(), portal.RequestIdKey, headers.RequestID)

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
