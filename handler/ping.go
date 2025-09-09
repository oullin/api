package handler

import (
	baseHttp "net/http"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/http"
)

type PingHandler struct{}

func MakePingHandler() PingHandler {
	return PingHandler{}
}

func (h PingHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	resp := http.MakeResponseFrom("0.0.1", w, r)
	now := time.Now().UTC()
	data := payload.PingResponse{
		Message: "pong",
		Date:    now.Format("2006-01-02"),
		Time:    now.Format("15:04:05"),
	}
	if err := resp.RespondOk(data); err != nil {
		return http.LogInternalError("could not encode ping response", err)
	}
	return nil
}
