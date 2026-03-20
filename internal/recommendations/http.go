package recommendations

import (
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/oullin/internal/shared/endpoint"
	"github.com/oullin/internal/shared/portal"
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
	data, err := portal.ParseJsonFile[RecommendationsResponse](h.filePath)

	if err != nil {
		slog.Error("Error reading recommendations file", "error", err)

		return endpoint.InternalError("could not read recommendations data")
	}

	data.Data = h.featured(data.Data)

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

func (h RecommendationsHandler) featured(items []RecommendationsData) []RecommendationsData {
	filtered := make([]RecommendationsData, 0, len(items))

	for _, item := range items {
		if item.Featured == 1 {
			filtered = append(filtered, item)
		}
	}

	type sortKey struct {
		time time.Time
		ok   bool
	}

	keys := make([]sortKey, len(filtered))
	for i, item := range filtered {
		t, ok := h.createdAt(item)
		keys[i] = sortKey{time: t, ok: ok}
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		left, right := keys[i], keys[j]

		switch {
		case left.ok && right.ok:
			return left.time.After(right.time)
		case left.ok:
			return true
		case right.ok:
			return false
		default:
			return false
		}
	})

	return filtered
}

func (h RecommendationsHandler) createdAt(item RecommendationsData) (time.Time, bool) {
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
