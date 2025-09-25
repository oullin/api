package handler

import (
	"fmt"
	baseHttp "net/http"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/portal"
)

type KeepAliveHandler struct {
	env *env.PingEnvironment
}

func MakeKeepAliveHandler(e *env.PingEnvironment) KeepAliveHandler {
	return KeepAliveHandler{env: e}
}

func (h KeepAliveHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	user, pass, ok := r.BasicAuth()

	if !ok || h.env.HasInvalidCreds(user, pass) {
		return http.LogUnauthorisedError(
			"invalid credentials",
			fmt.Errorf("invalid credentials"),
		)
	}

	resp := http.MakeNoCacheResponse(w, r)
	now := time.Now().UTC()

	data := payload.KeepAliveResponse{
		Message:  "pong",
		DateTime: now.Format(portal.DatesLayout),
	}

	if err := resp.RespondOk(data); err != nil {
		return http.LogInternalError("could not encode keep-alive response", err)
	}

	return nil
}
