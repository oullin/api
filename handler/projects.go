package handler

import (
	"github.com/oullin/handler/paginate"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
)

type ProjectsHandler struct {
	filePath            string
	cacheEnabled        bool
	publishedAtResolver ProjectsPublishedAtResolver
}

func NewProjectsHandler(filePath string) ProjectsHandler {
	return NewProjectsHandlerWithCache(filePath, true)
}

func NewProjectsHandlerWithCache(filePath string, cacheEnabled bool) ProjectsHandler {
	return NewProjectsHandlerWithResolver(filePath, cacheEnabled, nil)
}

func (h ProjectsHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.ProjectsResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading projects file", "error", err)

		return endpoint.InternalError("could not read projects data")
	}

	enrichProjectsResponse(r.Context(), &data, h.publishedAtResolver)

	page := paginate.NewFrom(r.URL, projectsPageSize)
	data = paginateProjectsResponse(data, page)

	resp, err := endpoint.NewResponseForPayload(data, 3600, h.cacheEnabled, w, r)
	if err != nil {
		slog.Error("Error preparing projects response cache", "error", err)

		return endpoint.InternalError("could not prepare projects response")
	}

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
