package middleware

import (
	"context"
	"crypto/subtle"
	"log/slog"
	baseHttp "net/http"
	"strings"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http"
)

const tokenHeader = "X-API-Key"
const usernameHeader = "X-API-Username"
const signatureHeader = "X-API-Signature"
const requestIDHeader = "X-Request-ID"

// Context keys for propagating auth info downstream
// Use unexported custom type to avoid collisions

type contextKey string

const (
	authAccountNameKey contextKey = "auth.account_name"
	requestIdKey       contextKey = "request.id"
)

type TokenCheckMiddleware struct {
	ApiKeys      *repository.ApiKeys
	TokenHandler *auth.TokenHandler
}

func MakeTokenMiddleware(tokenHandler *auth.TokenHandler, apiKeys *repository.ApiKeys) TokenCheckMiddleware {
	return TokenCheckMiddleware{
		ApiKeys:      apiKeys,
		TokenHandler: tokenHandler,
	}
}

func (t TokenCheckMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		reqID := strings.TrimSpace(r.Header.Get(requestIDHeader))

		if reqID == "" {
			return t.getInvalidRequestError()
		}

		logger := slog.With("request_id", reqID, "path", r.URL.Path, "method", r.Method)

		accountName := strings.TrimSpace(r.Header.Get(usernameHeader))
		publicToken := strings.TrimSpace(r.Header.Get(tokenHeader))
		signature := strings.TrimSpace(r.Header.Get(signatureHeader))

		if accountName == "" || publicToken == "" || signature == "" {
			logger.Warn("missing authentication headers")
			return t.getInvalidRequestError()
		}

		if err := auth.ValidateTokenFormat(publicToken); err != nil {
			logger.Warn("invalid token format")
			return t.getInvalidTokenFormatError()
		}

		reject := t.shallReject(logger, accountName, publicToken, signature)
		if reject {
			return t.getUnauthenticatedError()
		}

		// Update the request context
		ctx := context.WithValue(r.Context(), authAccountNameKey, accountName)
		ctx = context.WithValue(r.Context(), requestIdKey, reqID)
		r = r.WithContext(ctx)

		logger.Info("authentication successful")

		return next(w, r)
	}
}

func (t TokenCheckMiddleware) shallReject(logger *slog.Logger, accountName, publicToken, signature string) bool {
	var item *database.APIKey

	if item = t.ApiKeys.FindBy(accountName); item == nil {
		logger.Warn("account not found")
		return true
	}

	token, err := t.TokenHandler.DecodeTokensFor(
		item.AccountName,
		item.SecretKey,
		item.PublicKey,
	)

	if err != nil {
		logger.Error("failed to decode account keys", "account", item.AccountName, "error", err)
		return true
	}

	// Constant-time compare of provided public token vs stored one
	provided := []byte(strings.TrimSpace(publicToken))
	expected := []byte(strings.TrimSpace(token.PublicKey))
	if subtle.ConstantTimeCompare(provided, expected) != 1 {
		logger.Warn("public token mismatch", "account", item.AccountName)
		return true
	}

	// Compute local signature and compare in constant time
	localSignature := auth.CreateSignatureFrom(token.AccountName, token.SecretKey)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(localSignature)) != 1 {
		logger.Warn("signature mismatch", "account", item.AccountName)
		return true
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
