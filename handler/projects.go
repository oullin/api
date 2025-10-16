package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	baseHttp "net/http"
)

type ProjectsHandler struct {
	filePath string
}

func MakeProjectsHandler(filePath string) ProjectsHandler {
	return ProjectsHandler{
		filePath: filePath,
	}
}

func (h ProjectsHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.ProjectsResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading projects file", "error", err)

		return endpoint.InternalError("could not read projects data")
	}

	resp := endpoint.MakeResponseFrom(data.Version, w, r)

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for projects response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
