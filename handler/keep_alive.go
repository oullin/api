package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/http"
	baseHttp "net/http"
)

type KeepAliveHandler struct{}

func MakeKeepAliveHandler() KeepAliveHandler {
	return KeepAliveHandler{}
}

func (h KeepAliveHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	resp := http.MakeResponseFrom("0.0.1", w, r)
	data := payload.KeepAliveResponse{Message: "alive"}
	if err := resp.RespondOk(data); err != nil {
		return http.LogInternalError("could not encode keep alive response", err)
	}
	return nil
}
