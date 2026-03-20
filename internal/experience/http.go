package experience

import (
	"log/slog"
	"net/http"

	"github.com/oullin/internal/shared/endpoint"
	"github.com/oullin/internal/shared/portal"
)

type ExperienceHandler struct {
	filePath     string
	cacheEnabled bool
}

func NewExperienceHandler(filePath string) ExperienceHandler {
	return NewExperienceHandlerWithCache(filePath, true)
}

func NewExperienceHandlerWithCache(filePath string, cacheEnabled bool) ExperienceHandler {
	return ExperienceHandler{
		filePath:     filePath,
		cacheEnabled: cacheEnabled,
	}
}

func (h ExperienceHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[ExperienceResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading experience file", "error", err)

		return endpoint.InternalError("could not read experience data")
	}

	resp, err := endpoint.NewResponseForPayload(data, 3600, h.cacheEnabled, w, r)
	if err != nil {
		slog.Error("Error preparing experience response cache", "error", err)

		return endpoint.InternalError("could not prepare experience response")
	}

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for experience response", "error", err)

		return endpoint.InternalError("could not encode experience response")
	}

	return nil // A nil return indicates success.
}
