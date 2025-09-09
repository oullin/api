package handler

import (
	baseHttp "net/http"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/http"
)

type PingHandler struct {
	username string
	password string
}

func MakePingHandler() PingHandler {
	return PingHandler{
		username: env.GetSecretOrEnv("ping_username", "PING_USERNAME"),
		password: env.GetSecretOrEnv("ping_password", "PING_PASSWORD"),
	}
}

func (h PingHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	user, pass, ok := r.BasicAuth()
	if !ok || user != h.username || pass != h.password {
		return &http.ApiError{Message: "Unauthorized", Status: baseHttp.StatusUnauthorized}
	}

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
