package handler

import (
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"os"
)

type ExperienceHandler struct {
	content string
}

func MakeExperienceHandler() ExperienceHandler {
	return ExperienceHandler{
		content: "./storage/fixture/experience.json",
	}
}

func (h ExperienceHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	fixture, err := os.ReadFile(h.content)

	if err != nil {
		slog.Error("Error reading experience file: %v", err)

		return http.InternalError("could not read experience data")
	}

	if err := writeJSON(fixture, w); err != nil {
		return http.InternalError(err.Error())
	}

	return nil // A nil return indicates success.
}
