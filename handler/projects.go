package handler

import (
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
	"os"
)

type ProjectsHandler struct {
	content string
}

func MakeProjectsHandler() ProjectsHandler {
	return ProjectsHandler{
		content: "./storage/fixture/projects.json",
	}
}

func (h ProjectsHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	fixture, err := os.ReadFile(h.content)

	if err != nil {
		slog.Error("Error reading projects file", "error", err)

		return http.InternalError("could not read projects data")
	}

	if err := writeJSON(fixture, w); err != nil {
		return http.InternalError(err.Error())
	}

	return nil // A nil return indicates success.
}
