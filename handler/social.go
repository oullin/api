package handler

import (
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"os"
)

type SocialHandler struct {
	content string
}

func MakeSocialHandler() SocialHandler {
	return SocialHandler{
		content: "./storage/fixture/social.json",
	}
}

func (h SocialHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	fixture, err := os.ReadFile(h.content)

	if err != nil {
		slog.Error("Error reading social file: %v", err)

		return http.InternalError("could not read social data")
	}

	if err := writeJSON(fixture, w); err != nil {
		return http.InternalError(err.Error())
	}

	return nil // A nil return indicates success.
}
