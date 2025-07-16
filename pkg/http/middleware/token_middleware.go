package middleware

import (
	"fmt"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"strings"
)

const tokenHeader = "X-API-Key"
const usernameHeader = "X-API-Username"

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

		if accountName == "" || publicToken == "" {
			return &http.ApiError{
				Message: fmt.Sprintf("invalid request. Please, provide a valid token and accout name"),
				Status:  baseHttp.StatusForbidden,
			}
		}

		validPublicToken, err := auth.ValidateBearerToken(publicToken)
		if err != nil {
			return &http.ApiError{
				Message: fmt.Sprintf("invalid token format: [token: %s]", publicToken),
				Status:  baseHttp.StatusForbidden,
			}
		}

		item := t.ApiKeys.FindBy(accountName)
		if item == nil {
			return &http.ApiError{
				Message: fmt.Sprintf("the provided account does not exist: [account: %s]", accountName),
				Status:  baseHttp.StatusForbidden,
			}
		}

		if item.PublicKey != validPublicToken.Token {
			return &http.ApiError{
				Message: fmt.Sprintf("the provided token does not match your provided account: [token: %s, account name: %s]", publicToken, accountName),
				Status:  baseHttp.StatusForbidden,
			}
		}

		slog.Info("Token validation successful")

		return next(w, r)
	}
}
