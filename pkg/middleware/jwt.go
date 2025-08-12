package middleware

import (
	"context"
	baseHttp "net/http"
	"strings"

	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http"
)

type jwtContextKey string

const JWTClaimsKey jwtContextKey = "jwt.claims"

// JWTMiddleware validates Authorization Bearer tokens and injects claims into the request context.
type JWTMiddleware struct {
	Handler auth.JWTHandler
}

// Handle checks the Authorization header for a valid JWT token.
func (m JWTMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			return &http.ApiError{Message: "missing or invalid authorization header", Status: baseHttp.StatusUnauthorized}
		}

		tokenStr := strings.TrimSpace(header[len("bearer "):])
		claims, err := m.Handler.Validate(tokenStr)
		if err != nil {
			return &http.ApiError{Message: "invalid token", Status: baseHttp.StatusUnauthorized}
		}

		ctx := context.WithValue(r.Context(), JWTClaimsKey, claims)
		return next(w, r.WithContext(ctx))
	}
}

// GetJWTClaims extracts JWT claims from the context.
func GetJWTClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(JWTClaimsKey).(*auth.Claims)
	return claims, ok
}
