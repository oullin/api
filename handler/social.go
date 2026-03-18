package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
)

type LinksHandler struct {
	filePath     string
	cacheEnabled bool
}

type SocialHandler = LinksHandler

func NewLinksHandler(filePath string) LinksHandler {
	return NewLinksHandlerWithCache(filePath, true)
}

func NewLinksHandlerWithCache(filePath string, cacheEnabled bool) LinksHandler {
	return LinksHandler{
		filePath:     filePath,
		cacheEnabled: cacheEnabled,
	}
}

func NewSocialHandler(filePath string) LinksHandler {
	return NewLinksHandler(filePath)
}

func NewSocialHandlerWithCache(filePath string, cacheEnabled bool) LinksHandler {
	return NewLinksHandlerWithCache(filePath, cacheEnabled)
}

func (h LinksHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.LinksResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading links file", "error", err)

		return endpoint.InternalError("could not read links data")
	}

	// Cache for 1 week (604800 seconds) since links data rarely changes.
	resp, err := endpoint.NewResponseForPayload(data, 604800, h.cacheEnabled, w, r)
	if err != nil {
		slog.Error("Error preparing links response cache", "error", err)

		return endpoint.InternalError("could not prepare links response")
	}

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for links response", "error", err)

		return endpoint.InternalError("could not encode links response")
	}

	return nil // A nil return indicates success.
}
