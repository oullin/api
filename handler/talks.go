package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
)

type TalksHandler struct {
	filePath     string
	cacheEnabled bool
}

func NewTalksHandler(filePath string) TalksHandler {
	return NewTalksHandlerWithCache(filePath, true)
}

func NewTalksHandlerWithCache(filePath string, cacheEnabled bool) TalksHandler {
	return TalksHandler{
		filePath:     filePath,
		cacheEnabled: cacheEnabled,
	}
}

func (h TalksHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.TalksResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading talks file", "error", err)

		return endpoint.InternalError("could not read talks data")
	}

	resp, err := endpoint.NewResponseForPayload(data, 3600, h.cacheEnabled, w, r)
	if err != nil {
		slog.Error("Error preparing talks response cache", "error", err)

		return endpoint.InternalError("could not prepare talks response")
	}

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for talks response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
