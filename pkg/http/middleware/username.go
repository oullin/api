package middleware

import (
	"fmt"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"strings"
)

const usernameHeader = "X-API-Username"

func UsernameCheck(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		username := strings.TrimSpace(r.Header.Get(usernameHeader))

		if username != "gocanto" {
			message := fmt.Sprintf("Unauthorized: Invalid API username received ('%s')", username)
			slog.Error(message)

			return &http.ApiError{
				Message: message,
				Status:  baseHttp.StatusUnauthorized,
			}
		}

		slog.Info("Successfully authenticated user: gocanto")

		return next(w, r)
	}
}
