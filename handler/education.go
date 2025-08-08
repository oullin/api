package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/portal"

	"log/slog"
	baseHttp "net/http"
)

type EducationHandler struct {
	filePath string
}

func MakeEducationHandler(filePath string) EducationHandler {
	return EducationHandler{
		filePath: filePath,
	}
}

func (h EducationHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	data, err := portal.ParseJsonFile[payload.EducationResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading education file", "error", err)

		return http.InternalError("could not read education data")
	}

	resp := http.MakeResponseFrom(data.Version, w, r)

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
