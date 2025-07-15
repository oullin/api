package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/http"
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

func (h RecommendationsHandler) Handle(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
	data, err := pkg.ParseJsonFile[payload.TalksResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading recommendations file", "error", err)

		return http.InternalError("could not read recommendations data")
	}

	resp := http.MakeResponseFrom(data.Version, w, r)

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
