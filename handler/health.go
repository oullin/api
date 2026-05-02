package handler

import (
	"net/http"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"
)

type HealthHandler struct{}

func NewHealthHandler() HealthHandler {
	return HealthHandler{}
}

func (h HealthHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	resp := endpoint.NewNoCacheResponse(w, r)

	data := payload.KeepAliveResponse{
		Message:  "ok",
		DateTime: time.Now().UTC().Format(portal.DatesLayout),
	}

	if err := resp.RespondOk(data); err != nil {
		return endpoint.LogInternalError("could not encode health response", err)
	}

	return nil
}
