package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/http"
	"log/slog"
	baseHttp "net/http"
)

type ExperienceHandler struct {
	filePath string
}

func MakeExperienceHandler(filePath string) ExperienceHandler {
	return ExperienceHandler{
		filePath: filePath,
	}
}

func (h ExperienceHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	data, err := pkg.ParseJsonFile[payload.ExperienceResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading experience file", "error", err)

		return http.InternalError("could not read experience data")
	}

	resp := http.MakeResponseFrom(data.Version, w, r)

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
