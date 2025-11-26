package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
)

type SocialHandler struct {
	filePath string
}

func NewSocialHandler(filePath string) SocialHandler {
	return SocialHandler{
		filePath: filePath,
	}
}

func (h SocialHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.SocialResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading social file", "error", err)

		return endpoint.InternalError("could not read social data")
	}

	// Cache for 1 week (604800 seconds) since social data rarely changes
	resp := endpoint.NewResponseWithCache(data.Version, 604800, w, r)

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for social response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
