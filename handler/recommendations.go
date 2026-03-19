package handler

import (
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"

	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"
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

	data.Data = featuredRecommendations(data.Data)

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

func featuredRecommendations(items []payload.RecommendationsData) []payload.RecommendationsData {
	filtered := make([]payload.RecommendationsData, 0, len(items))

	for _, item := range items {
		if item.Featured == 1 {
			filtered = append(filtered, item)
		}
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Featured != filtered[j].Featured {
			return filtered[i].Featured > filtered[j].Featured
		}

		left, leftOK := recommendationCreatedAt(filtered[i])
		right, rightOK := recommendationCreatedAt(filtered[j])

		switch {
		case leftOK && rightOK:
			return left.After(right)
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return false
		}
	})

	return filtered
}

func recommendationCreatedAt(item payload.RecommendationsData) (time.Time, bool) {
	createdAt := strings.TrimSpace(item.CreatedAt)
	if createdAt == "" {
		return time.Time{}, false
	}

	parsed, err := time.Parse("2006-01-02", createdAt)
	if err != nil {
		return time.Time{}, false
	}

	return parsed, true
}
