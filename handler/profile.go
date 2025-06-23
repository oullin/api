package handler

import (
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"os"
)

type ProfileHandler struct {
	content string
}

func MakeProfileHandler() ProfileHandler {
	return ProfileHandler{
		content: "./storage/fixture/profile.json",
	}
}

func (h ProfileHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	fixture, err := os.ReadFile(h.content)

	if err != nil {
		slog.Error("Error reading profile file: %v", err)

		return http.InternalError("could not read profile data")
	}

	if err := writeJSON(fixture, w); err != nil {
		return http.InternalError(err.Error())
	}

	return nil // A nil return indicates success.
}
