package handler

import (
	"encoding/json"
	"log/slog"

	"github.com/oullin/database/repository"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/auth"
	pkgHttp "github.com/oullin/pkg/http"
	"github.com/oullin/pkg/portal"

	baseHttp "net/http"
)

// AuthHandler provides endpoints to generate JWT bearer tokens.
//
// It validates credentials against the api_keys repository and
// issues signed tokens using the provided JWT handler.
type AuthHandler struct {
	ApiKeys *repository.ApiKeys
	JWT     auth.JWTHandler
}

func MakeAuthHandler(repo *repository.ApiKeys, jwt auth.JWTHandler) AuthHandler {
	return AuthHandler{ApiKeys: repo, JWT: jwt}
}

// Token validates account credentials and returns a signed JWT.
func (h *AuthHandler) Token(w baseHttp.ResponseWriter, r *baseHttp.Request) *pkgHttp.ApiError {
	defer portal.CloseWithLog(r.Body)

	req, err := pkgHttp.ParseRequestBody[payload.TokenRequest](r)
	if err != nil {
		slog.Error("failed to parse request body", "err", err)
		return pkgHttp.BadRequestError("invalid request body")
	}
	if req.AccountName == "" || req.SecretKey == "" {
		return pkgHttp.BadRequestError("account_name and secret_key are required")
	}

	key := h.ApiKeys.FindBy(req.AccountName)
	if key == nil || string(key.SecretKey) != req.SecretKey {
		return &pkgHttp.ApiError{Message: "invalid credentials", Status: baseHttp.StatusUnauthorized}
	}

	token, err := h.JWT.Generate(req.AccountName)
	if err != nil {
		slog.Error("failed to generate token", "err", err)
		return pkgHttp.InternalError("could not generate token")
	}

	resp := payload.TokenResponse{Token: token}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode response", "err", err)
		return pkgHttp.InternalError("There was an issue processing the response. Please, try later.")
	}

	return nil
}
