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

type PingHandler struct {
	env *env.Ping
}

func MakePingHandler(e *env.Ping) PingHandler {
	return PingHandler{env: e}
}

func (h PingHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	user, pass, ok := r.BasicAuth()

	if !ok || h.env.HasInvalidCreds(user, pass) {
		return http.LogUnauthorisedError(
			"invalid credentials",
			fmt.Errorf("invalid credentials"),
		)
	}

	resp := http.MakeResponseFrom("0.0.1", w, r)
	now := time.Now().UTC()

	data := payload.PingResponse{
		Message:  "pong",
		DateTime: now.Format(portal.DatesLayout),
	}

	if err := resp.RespondOk(data); err != nil {
		return http.LogInternalError("could not encode ping response", err)
	}

	return nil
}
