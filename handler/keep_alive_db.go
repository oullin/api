package handler

import (
	"fmt"
	baseHttp "net/http"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"
)

type KeepAliveDBHandler struct {
	env *env.PingEnvironment
	db  *database.Connection
}

func MakeKeepAliveDBHandler(e *env.PingEnvironment, db *database.Connection) KeepAliveDBHandler {
	return KeepAliveDBHandler{env: e, db: db}
}

func (h KeepAliveDBHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *endpoint.ApiError {
	user, pass, ok := r.BasicAuth()

	if !ok || h.env.HasInvalidCreds(user, pass) {
		return endpoint.LogUnauthorisedError(
			"invalid credentials",
			fmt.Errorf("invalid credentials"),
		)
	}

	if err := h.db.Ping(); err != nil {
		return endpoint.LogInternalError("database ping failed", err)
	}

	resp := endpoint.MakeNoCacheResponse(w, r)
	now := time.Now().UTC()

	data := payload.KeepAliveResponse{
		Message:  "pong",
		DateTime: now.Format(portal.DatesLayout),
	}

	if err := resp.RespondOk(data); err != nil {
		return endpoint.LogInternalError("could not encode keep-alive response", err)
	}

	return nil
}
