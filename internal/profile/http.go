package profile

import (
	"log/slog"
	"net/http"

	"github.com/oullin/internal/shared/endpoint"
	"github.com/oullin/internal/shared/portal"
)

type ProfileHandler struct {
	filePath     string
	cacheEnabled bool
}

func NewProfileHandler(filePath string) ProfileHandler {
	return NewProfileHandlerWithCache(filePath, true)
}

func NewProfileHandlerWithCache(filePath string, cacheEnabled bool) ProfileHandler {
	return ProfileHandler{
		filePath:     filePath,
		cacheEnabled: cacheEnabled,
	}
}

func (h ProfileHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[ProfileResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading profile file", "error", err)

		return endpoint.InternalError("could not read profile data")
	}

	resp, err := endpoint.NewResponseForPayload(data, 3600, h.cacheEnabled, w, r)
	if err != nil {
		slog.Error("Error preparing profile response cache", "error", err)

		return endpoint.InternalError("could not prepare profile response")
	}

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for profile response", "error", err)

		return endpoint.InternalError("could not encode profile response")
	}

	return nil // A nil return indicates success.
}
