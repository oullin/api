package middleware

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/repoentity"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cache"
	"github.com/oullin/pkg/endpoint"
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

func NewTokenMiddleware(tokenHandler *auth.TokenHandler, apiKeys *repository.ApiKeys) TokenCheckMiddleware {
	return TokenCheckMiddleware{
		maxFailPerScope: 50,
		disallowFuture:  true,
		ApiKeys:         apiKeys,
		now:             time.Now,
		TokenHandler:    tokenHandler,
		clockSkew:       10 * time.Minute,
		failWindow:      5 * time.Minute,
		nonceTTL:        25 * time.Minute,
		nonceCache:      cache.NewTTLCache(),
		rateLimiter:     limiter.NewMemoryLimiter(5*time.Minute, 50),
	}
}

func (t TokenCheckMiddleware) Handle(next endpoint.ApiHandler) endpoint.ApiHandler {
	return func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
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
		vt := mwguards.NewValidTimestamp(headers.Timestamp, t.now)
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

func (t TokenCheckMiddleware) GuardDependencies() *endpoint.ApiError {
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

func (t TokenCheckMiddleware) ValidateAndGetHeaders(r *http.Request, requestId string) (AuthTokenHeaders, *endpoint.ApiError) {
	intendedOriginURL := portal.IntendedOriginFromHeader(r.Header)
	accountName := strings.TrimSpace(r.Header.Get(portal.UsernameHeader))
	signature := strings.TrimSpace(r.Header.Get(portal.SignatureHeader))
	publicToken := strings.TrimSpace(r.Header.Get(portal.TokenHeader))
	ts := strings.TrimSpace(r.Header.Get(portal.TimestampHeader))
	nonce := strings.TrimSpace(r.Header.Get(portal.NonceHeader))
	ip := portal.ParseClientIP(r)

	missing := []string{}
	if accountName == "" {
		missing = append(missing, portal.UsernameHeader)
	}

	if publicToken == "" {
		missing = append(missing, portal.TokenHeader)
	}

	if signature == "" {
		missing = append(missing, portal.SignatureHeader)
	}

	if ts == "" {
		missing = append(missing, portal.TimestampHeader)
	}

	if nonce == "" {
		missing = append(missing, portal.NonceHeader)
	}

	if ip == "" {
		missing = append(missing, "client_ip")
	}

	if intendedOriginURL == "" {
		missing = append(missing, "intended_origin")
	}

	if len(missing) > 0 {
		message := "Invalid authentication headers: missing " + strings.Join(missing, ", ")

		return AuthTokenHeaders{}, mwguards.InvalidRequestError(
			message,
			message,
			map[string]any{"missing": missing, "request_id": requestId},
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

func (t TokenCheckMiddleware) AttachContext(r *http.Request, headers AuthTokenHeaders) *http.Request {
	ctx := context.WithValue(r.Context(), portal.AuthAccountNameKey, headers.AccountName)
	ctx = context.WithValue(ctx, portal.RequestIDKey, headers.RequestID)

	return r.WithContext(ctx)
}

func (t TokenCheckMiddleware) HasInvalidFormat(headers AuthTokenHeaders) (*database.APIKey, *endpoint.ApiError) {
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
		t.rateLimiter.Fail(limiterKey)

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

func (t TokenCheckMiddleware) buildAuthErrorContext(headers AuthTokenHeaders, includeNonce bool) map[string]any {
	ctx := map[string]any{
		"account_name": headers.AccountName,
		"client_ip":    headers.ClientIP,
		"origin":       headers.IntendedOriginURL,
		"request_id":   headers.RequestID,
		"timestamp":    headers.Timestamp,
	}

	if includeNonce && headers.Nonce != "" {
		// Include first 8 chars of nonce for debugging (safe for logs)
		// If nonce is shorter, include what we have
		nonceLen := len(headers.Nonce)
		if nonceLen > 8 {
			ctx["nonce"] = headers.Nonce[:8] + "..."
		} else {
			ctx["nonce"] = headers.Nonce
		}
	}

	return ctx
}

func (t TokenCheckMiddleware) HasInvalidSignature(headers AuthTokenHeaders, apiKey *database.APIKey) *endpoint.ApiError {
	var err error
	var byteSignature []byte
	limiterKey := headers.ClientIP + "|" + strings.ToLower(headers.AccountName)

	if byteSignature, err = hex.DecodeString(headers.Signature); err != nil {
		t.rateLimiter.Fail(limiterKey)

		return mwguards.UnauthenticatedError(
			"Invalid signature format",
			"error decoding signature string: "+err.Error(),
			t.buildAuthErrorContext(headers, false),
		)
	}

	entity := repoentity.FindSignatureFrom{
		Key:        apiKey,
		Signature:  byteSignature,
		Origin:     portal.NormalizeOriginWithPath(headers.IntendedOriginURL),
		ServerTime: t.now(),
	}

	signature := t.ApiKeys.FindSignatureFrom(entity)

	if signature == nil {
		t.rateLimiter.Fail(limiterKey)

		return mwguards.UnauthenticatedError(
			"Invalid signature",
			"signature not found",
			t.buildAuthErrorContext(headers, true),
		)
	}

	if err = t.ApiKeys.IncreaseSignatureTries(signature.UUID, signature.CurrentTries+1); err != nil {
		t.rateLimiter.Fail(limiterKey)

		return mwguards.InvalidRequestError("could not increase signature tries", err.Error())
	}

	return nil
}
