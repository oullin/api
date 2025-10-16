package handler

import (
	"fmt"
	baseHttp "net/http"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"
)

type KeepAliveHandler struct {
	env *env.PingEnvironment
}

func MakeKeepAliveHandler(e *env.PingEnvironment) KeepAliveHandler {
	return KeepAliveHandler{env: e}
}

func (h KeepAliveHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *endpoint.ApiError {
	user, pass, ok := r.BasicAuth()

	if !ok || h.env.HasInvalidCreds(user, pass) {
		return endpoint.LogUnauthorisedError(
			"invalid credentials",
			fmt.Errorf("invalid credentials"),
		)
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
