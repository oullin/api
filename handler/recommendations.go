package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
)

type RecommendationsHandler struct {
	filePath     string
	cacheEnabled bool
}

func NewRecommendationsHandler(filePath string) RecommendationsHandler {
	return NewRecommendationsHandlerWithCache(filePath, true)
}

func NewRecommendationsHandlerWithCache(filePath string, cacheEnabled bool) RecommendationsHandler {
	return RecommendationsHandler{
		filePath:     filePath,
		cacheEnabled: cacheEnabled,
	}
}

func (h RecommendationsHandler) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.RecommendationsResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading recommendations file", "error", err)

		return endpoint.InternalError("could not read recommendations data")
	}

	resp, err := endpoint.NewResponseForPayload(data, 3600, h.cacheEnabled, w, r)
	if err != nil {
		slog.Error("Error preparing recommendations response cache", "error", err)

		return endpoint.InternalError("could not prepare recommendations response")
	}

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for recommendations response", "error", err)

		return endpoint.InternalError("could not encode recommendations response")
	}

	return nil // A nil return indicates success.
}
