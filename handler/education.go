package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
)

type EducationHandler struct {
	filePath string
}

func NewEducationHandler(filePath string) EducationHandler {
	return EducationHandler{
		filePath: filePath,
	}
}

func (h EducationHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.EducationResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading education file", "error", err)

		return endpoint.InternalError("could not read education data")
	}

	resp := endpoint.NewResponseFrom(data.Version, w, r)

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for education response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
