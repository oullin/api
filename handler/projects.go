package handler

import (
	"log/slog"
	"net/http"

	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler/paginate"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"
	"github.com/oullin/pkg/projects"
)

type ProjectsHandler struct {
	filePath     string
	cacheEnabled bool
}

func NewProjectsHandler(filePath string) ProjectsHandler {
	return NewProjectsHandlerWithCache(filePath, true)
}

func NewProjectsHandlerWithCache(filePath string, cacheEnabled bool) ProjectsHandler {
	return ProjectsHandler{
		filePath:     filePath,
		cacheEnabled: cacheEnabled,
	}
}

func (h ProjectsHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.ProjectsResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading projects file", "error", err)

		return endpoint.InternalError("could not read projects data")
	}

	projects.EnrichResponse(&data)

	page := paginate.NewFrom(r.URL, projects.PageSize)
	page.SetNumItems(int64(len(data.Data)))

	start := (page.Page - 1) * page.Limit
	switch {
	case start < 0:
		start = 0
	case start > len(data.Data):
		start = len(data.Data)
	}

	end := start + page.Limit
	if end > len(data.Data) {
		end = len(data.Data)
	}

	items := append([]payload.ProjectsData(nil), data.Data[start:end]...)
	result := pagination.NewPagination(items, page)
	response := payload.ProjectsResponse{
		Version:      data.Version,
		Data:         result.Data,
		Page:         result.Page,
		Total:        result.Total,
		PageSize:     result.PageSize,
		TotalPages:   result.TotalPages,
		NextPage:     result.NextPage,
		PreviousPage: result.PreviousPage,
	}

	resp, err := endpoint.NewResponseForPayload(response, 3600, h.cacheEnabled, w, r)
	if err != nil {
		slog.Error("Error preparing projects response cache", "error", err)

		return endpoint.InternalError("could not prepare projects response")
	}

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(response); err != nil {
		slog.Error("Error marshaling JSON for projects response", "error", err)

		return endpoint.InternalError("could not encode projects response")
	}

	return nil // A nil return indicates success.
}
