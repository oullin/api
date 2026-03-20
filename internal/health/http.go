package health

import (
	"fmt"
	"net/http"
	"time"

	env "github.com/oullin/internal/app/config"
	"github.com/oullin/internal/shared/endpoint"
	"github.com/oullin/internal/shared/portal"
)

type KeepAliveHandler struct {
	env *env.PingEnvironment
}

func NewKeepAliveHandler(e *env.PingEnvironment) KeepAliveHandler {
	return KeepAliveHandler{env: e}
}

func (h KeepAliveHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	user, pass, ok := r.BasicAuth()

	if !ok || h.env.HasInvalidCreds(user, pass) {
		return endpoint.LogUnauthorisedError(
			"invalid credentials",
			fmt.Errorf("invalid credentials"),
		)
	}

	resp := endpoint.NewNoCacheResponse(w, r)
	now := time.Now().UTC()

	data := KeepAliveResponse{
		Message:  "pong",
		DateTime: now.Format(portal.DatesLayout),
	}

	if err := resp.RespondOk(data); err != nil {
		return endpoint.LogInternalError("could not encode keep-alive response", err)
	}

	return nil
}
