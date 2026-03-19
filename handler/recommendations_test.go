package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/oullin/handler"
	"github.com/oullin/handler/payload"
)

func TestRecommendationsHandler(t *testing.T) {
	handler.RunFileHandlerTest(t, handler.FileHandlerTestCase{
		Make:     func(f string) handler.FileHandler { return handler.NewRecommendationsHandler(f) },
		Endpoint: "/recommendations",
		Fixture:  "../storage/fixture/recommendations.json",
		Assert:   handler.AssertEmptyData(),
	})
}

func TestRecommendationsHandlerFiltersAndSortsFeaturedItems(t *testing.T) {
	fixture := `{
		"version": "v1",
		"data": [
			{
				"uuid": "older-featured",
				"relation": "Worked together",
				"text": "Older featured",
				"featured": 1,
				"created_at": "2024-01-01",
				"updated_at": "2024-01-01",
				"person": {
					"avatar": "recommendation/older.jpeg",
					"full_name": "Older Featured",
					"company": "Example",
					"designation": "Engineer"
				}
			},
			{
				"uuid": "newer-not-featured",
				"relation": "Worked together",
				"text": "Should be excluded",
				"featured": 0,
				"created_at": "2025-01-01",
				"updated_at": "2025-01-01",
				"person": {
					"avatar": "recommendation/excluded.jpeg",
					"full_name": "Excluded",
					"company": "Example",
					"designation": "Engineer"
				}
			},
			{
				"uuid": "newer-featured",
				"relation": "Worked together",
				"text": "Newer featured",
				"featured": 1,
				"created_at": "2024-12-31",
				"updated_at": "2024-12-31",
				"person": {
					"avatar": "recommendation/newer.jpeg",
					"full_name": "Newer Featured",
					"company": "Example",
					"designation": "Engineer"
				}
			}
		]
	}`

	dir := t.TempDir()
	path := filepath.Join(dir, "recommendations.json")

	if err := os.WriteFile(path, []byte(fixture), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	h := handler.NewRecommendationsHandlerWithCache(path, false)
	req := httptest.NewRequest(http.MethodGet, "/recommendations", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var res payload.RecommendationsResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(res.Data) != 2 {
		t.Fatalf("expected 2 featured items, got %+v", res.Data)
	}

	if res.Data[0].UUID != "newer-featured" || res.Data[1].UUID != "older-featured" {
		t.Fatalf("unexpected order: %+v", res.Data)
	}

	if res.Data[0].Featured != 1 || res.Data[1].Featured != 1 {
		t.Fatalf("expected featured items only: %+v", res.Data)
	}
}
