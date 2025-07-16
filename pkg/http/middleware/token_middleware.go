package middleware

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"strings"
)

const tokenHeader = "X-API-Key"
const usernameHeader = "X-API-Username"
const signatureHeader = "X-API-Signature"

type TokenCheckMiddleware struct {
	ApiKeys *repository.ApiKeys
}

func MakeTokenMiddleware(apiKeys *repository.ApiKeys) TokenCheckMiddleware {
	return TokenCheckMiddleware{
		ApiKeys: apiKeys,
	}
}

func (t TokenCheckMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {

		accountName := strings.TrimSpace(r.Header.Get(usernameHeader))
		publicToken := strings.TrimSpace(r.Header.Get(tokenHeader))
		signature := strings.TrimSpace(r.Header.Get(signatureHeader))

		if accountName == "" || publicToken == "" || signature == "" {
			return t.getInvalidRequestError(accountName, publicToken, signature)
		}

		if err := auth.ValidateTokenFormat(publicToken); err != nil {
			return t.getInvalidTokenFormatError(publicToken, err)
		}

		if t.shallReject(accountName, signature) {
			return t.getUnauthenticatedError(accountName, publicToken, signature)
		}

		slog.Info("Token validation successful")

		return next(w, r)
	}
}

func (t TokenCheckMiddleware) shallReject(accountName, signature string) bool {
	var item *database.APIKey

	if item = t.ApiKeys.FindBy(accountName); item == nil {
		return true
	}

	token := auth.Token{
		AccountName: item.AccountName,
		SecretKey:   item.SecretKey,
		PublicKey:   item.PublicKey,
		Length:      len(item.PublicKey),
	}

	return token.HasInValidSignature(signature)
}

func (t TokenCheckMiddleware) getInvalidRequestError(accountName, publicToken, signature string) *http.ApiError {
	message := fmt.Sprintf(
		"invalid request. Please, provide a valid token, signature and accout name headers. [account: %s, public token: %s, signature: %s]",
		accountName,
		auth.SafeDisplay(publicToken),
		auth.SafeDisplay(signature),
	)

	return &http.ApiError{
		Message: message,
		Status:  baseHttp.StatusForbidden,
	}
}

func (t TokenCheckMiddleware) getInvalidTokenFormatError(publicToken string, err error) *http.ApiError {
	return &http.ApiError{
		Message: fmt.Sprintf("invalid token format [token: %s]: %v", auth.SafeDisplay(publicToken), err),
		Status:  baseHttp.StatusForbidden,
	}
}

func (t TokenCheckMiddleware) getUnauthenticatedError(accountName, publicToken, signature string) *http.ApiError {
	message := fmt.Sprintf(
		"Unauthenticated, please check your credentials and signature headers: [token: %s, account name: %s, signature: %s]",
		auth.SafeDisplay(publicToken),
		accountName,
		auth.SafeDisplay(signature),
	)

	return &http.ApiError{
		Message: message,
		Status:  baseHttp.StatusForbidden,
	}
}
