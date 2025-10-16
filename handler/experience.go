package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
)

type ExperienceHandler struct {
	filePath string
}

func NewExperienceHandler(filePath string) ExperienceHandler {
	return ExperienceHandler{
		filePath: filePath,
	}
}

func (h ExperienceHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.ExperienceResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading experience file", "error", err)

		return endpoint.InternalError("could not read experience data")
	}

	resp := endpoint.NewResponseFrom(data.Version, w, r)

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for experience response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
