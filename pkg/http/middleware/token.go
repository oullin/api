package middleware

import (
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
)

const tokenHeader = "X-API-Key"

type TokenCheckMiddleware struct {
	token auth.Token
}

func MakeTokenMiddleware(token auth.Token) TokenCheckMiddleware {
	return TokenCheckMiddleware{
		token: token,
	}
}

func (t TokenCheckMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {

		if t.token.IsInvalid(r.Header.Get(tokenHeader)) {
			message := "Forbidden: Invalid API key"
			slog.Error(message)

			return &http.ApiError{
				Message: message,
				Status:  baseHttp.StatusForbidden,
			}
		}

		slog.Info("Token validation successful")

		return next(w, r)
	}
}
