package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	baseHttp "net/http"
)

type RecommendationsHandler struct {
	filePath string
}

func MakeRecommendationsHandler(filePath string) RecommendationsHandler {
	return RecommendationsHandler{
		filePath: filePath,
	}
}

func (h RecommendationsHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *endpoint.ApiError {
	data, err := portal.ParseJsonFile[payload.RecommendationsResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading recommendations file", "error", err)

		return endpoint.InternalError("could not read recommendations data")
	}

	resp := endpoint.MakeResponseFrom(data.Version, w, r)

	if resp.HasCache() {
		resp.RespondWithNotModified()

		return nil
	}

	if err := resp.RespondOk(data); err != nil {
		slog.Error("Error marshaling JSON for recommendations response", "error", err)

		return nil
	}

	return nil // A nil return indicates success.
}
