package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/handler"
	"github.com/oullin/handler/payload"
)

func intPtr(v int) *int { return &v }

func TestProjectsHandler_SortsAndPaginates(t *testing.T) {
	fixture := writeProjectsFixture(t, payload.ProjectsResponse{
		Version: "1.0.0",
		Data: []payload.ProjectsData{
			{UUID: "project-1", Sort: intPtr(10), Title: "One", URL: "https://github.com/example/one", PublishedAt: "2026-03-01T00:00:00Z"},
			{UUID: "project-2", Sort: intPtr(1), Title: "Two", URL: "https://github.com/example/two", PublishedAt: "2026-03-10T00:00:00Z"},
			{UUID: "project-3", Sort: intPtr(9), Title: "Three", URL: "https://github.com/example/three", PublishedAt: "2026-03-03T00:00:00Z"},
			{UUID: "project-4", Sort: intPtr(3), Title: "Four", URL: "https://github.com/example/four", PublishedAt: "2026-03-08T00:00:00Z"},
			{UUID: "project-5", Sort: intPtr(8), Title: "Five", URL: "https://github.com/example/five", PublishedAt: "2026-03-02T00:00:00Z"},
			{UUID: "project-6", Sort: intPtr(4), Title: "Six", URL: "https://github.com/example/six", PublishedAt: "2026-03-07T00:00:00Z"},
			{UUID: "project-7", Sort: intPtr(5), Title: "Seven", URL: "https://github.com/example/seven", PublishedAt: "2026-03-06T00:00:00Z"},
			{UUID: "project-8", Sort: intPtr(2), Title: "Eight", URL: "https://github.com/example/eight", PublishedAt: "2026-03-09T00:00:00Z"},
			{UUID: "project-9", Sort: intPtr(6), Title: "Nine", URL: "https://github.com/example/nine", PublishedAt: "2026-03-05T00:00:00Z"},
			{UUID: "project-10", Sort: intPtr(7), Title: "Ten", URL: "https://github.com/example/ten", PublishedAt: "2026-03-04T00:00:00Z"},
		},
	})

	h := handler.NewProjectsHandlerWithCache(fixture, true)

	req := httptest.NewRequest(http.MethodGet, "/projects?page=1", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	var resp payload.ProjectsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Version != "1.0.0" {
		t.Fatalf("expected version to be preserved, got %q", resp.Version)
	}

	if resp.Page != 1 || resp.PageSize != 8 || resp.Total != 10 || resp.TotalPages != 2 {
		t.Fatalf("unexpected pagination: %+v", resp)
	}

	if resp.NextPage == nil || *resp.NextPage != 2 {
		t.Fatalf("expected next page to be 2, got %+v", resp.NextPage)
	}

	if resp.PreviousPage != nil {
		t.Fatalf("expected previous page to be nil, got %+v", resp.PreviousPage)
	}

	if len(resp.Data) != 8 {
		t.Fatalf("expected 8 items on page 1, got %d", len(resp.Data))
	}

	if resp.Data[0].UUID != "project-2" || resp.Data[0].Sort == nil || *resp.Data[0].Sort != 1 {
		t.Fatalf("expected lowest sort project first, got %+v", resp.Data[0])
	}

	if resp.Data[1].UUID != "project-8" {
		t.Fatalf("expected second-lowest sort project second, got %+v", resp.Data[1])
	}
}

func TestProjectsHandler_Page2(t *testing.T) {
	fixture := writeProjectsFixture(t, payload.ProjectsResponse{
		Version: "1.0.0",
		Data: []payload.ProjectsData{
			{UUID: "project-1", Sort: intPtr(9), Title: "One", URL: "https://github.com/example/one", PublishedAt: "2026-03-09T00:00:00Z"},
			{UUID: "project-2", Sort: intPtr(8), Title: "Two", URL: "https://github.com/example/two", PublishedAt: "2026-03-08T00:00:00Z"},
			{UUID: "project-3", Sort: intPtr(7), Title: "Three", URL: "https://github.com/example/three", PublishedAt: "2026-03-07T00:00:00Z"},
			{UUID: "project-4", Sort: intPtr(6), Title: "Four", URL: "https://github.com/example/four", PublishedAt: "2026-03-06T00:00:00Z"},
			{UUID: "project-5", Sort: intPtr(5), Title: "Five", URL: "https://github.com/example/five", PublishedAt: "2026-03-05T00:00:00Z"},
			{UUID: "project-6", Sort: intPtr(4), Title: "Six", URL: "https://github.com/example/six", PublishedAt: "2026-03-04T00:00:00Z"},
			{UUID: "project-7", Sort: intPtr(3), Title: "Seven", URL: "https://github.com/example/seven", PublishedAt: "2026-03-03T00:00:00Z"},
			{UUID: "project-8", Sort: intPtr(2), Title: "Eight", URL: "https://github.com/example/eight", PublishedAt: "2026-03-02T00:00:00Z"},
			{UUID: "project-9", Sort: intPtr(1), Title: "Nine", URL: "https://github.com/example/nine", PublishedAt: "2026-03-01T00:00:00Z"},
		},
	})

	h := handler.NewProjectsHandlerWithCache(fixture, true)

	req := httptest.NewRequest(http.MethodGet, "/projects?page=2", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	var resp payload.ProjectsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Page != 2 || resp.Total != 9 || resp.TotalPages != 2 {
		t.Fatalf("unexpected pagination: %+v", resp)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected one item on page 2, got %d", len(resp.Data))
	}

	if resp.PreviousPage == nil || *resp.PreviousPage != 1 {
		t.Fatalf("expected previous page to be 1, got %+v", resp.PreviousPage)
	}

	if resp.NextPage != nil {
		t.Fatalf("expected next page to be nil, got %+v", resp.NextPage)
	}
}

func TestProjectsHandler_SortsBySortWithPublishedAtTieBreaker(t *testing.T) {
	fixture := writeProjectsFixture(t, payload.ProjectsResponse{
		Version: "1.0.0",
		Data: []payload.ProjectsData{
			{
				UUID:        "project-older",
				Sort:        intPtr(1),
				Title:       "Older",
				URL:         "https://github.com/example/older",
				PublishedAt: "2026-03-01T00:00:00Z",
			},
			{
				UUID:        "project-newer",
				Sort:        intPtr(1),
				Title:       "Newer",
				URL:         "https://github.com/example/newer",
				PublishedAt: "2026-03-10T00:00:00Z",
			},
			{
				UUID:        "project-later-sort",
				Sort:        intPtr(2),
				Title:       "Later Sort",
				URL:         "https://github.com/example/later-sort",
				PublishedAt: "2026-03-17T12:00:00Z",
			},
		},
	})

	h := handler.NewProjectsHandlerWithCache(fixture, true)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	var resp payload.ProjectsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 items, got %d", len(resp.Data))
	}

	if resp.Data[0].UUID != "project-newer" {
		t.Fatalf("expected newer project first within equal sort, got %+v", resp.Data[0])
	}

	if resp.Data[1].UUID != "project-older" {
		t.Fatalf("expected older project second within equal sort, got %+v", resp.Data[1])
	}

	if resp.Data[2].UUID != "project-later-sort" {
		t.Fatalf("expected higher sort project last, got %+v", resp.Data[2])
	}
}

func TestProjectsHandler_NoStoreWhenCacheDisabled(t *testing.T) {
	fixture := writeProjectsFixture(t, payload.ProjectsResponse{
		Version: "1.0.0",
		Data: []payload.ProjectsData{
			{UUID: "project-1", Sort: intPtr(1), Title: "One", URL: "https://github.com/example/one", PublishedAt: "2026-03-01T00:00:00Z"},
		},
	})

	h := handler.NewProjectsHandlerWithCache(fixture, false)
	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("cache-control %q", rec.Header().Get("Cache-Control"))
	}

	if rec.Header().Get("ETag") != "" {
		t.Fatalf("expected empty etag, got %q", rec.Header().Get("ETag"))
	}
}

func TestProjectsFixture_ContainsValidPublishedAtAndSort(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("..", "storage", "fixture", "projects.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var raw struct {
		Data []map[string]any `json:"data"`
	}

	if err := json.Unmarshal(body, &raw); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}

	if len(raw.Data) == 0 {
		t.Fatalf("expected fixture data")
	}

	for _, item := range raw.Data {
		uuid, _ := item["uuid"].(string)

		publishedAt, ok := item["published_at"]
		if !ok {
			t.Fatalf("expected published_at in fixture item %s: %+v", uuid, item)
		}

		publishedAtStr, _ := publishedAt.(string)
		if strings.TrimSpace(publishedAtStr) == "" {
			t.Fatalf("expected non-empty published_at in fixture item %s", uuid)
		}

		sortVal, ok := item["sort"]
		if !ok {
			t.Fatalf("expected sort in fixture item %s: %+v", uuid, item)
		}

		sortNum, ok := sortVal.(float64)
		if !ok || sortNum <= 0 {
			t.Fatalf("expected positive sort in fixture item %s, got %v", uuid, sortVal)
		}

		if _, ok := item["created_at"]; ok {
			t.Fatalf("did not expect created_at in fixture item: %+v", item)
		}

		if _, ok := item["updated_at"]; ok {
			t.Fatalf("did not expect updated_at in fixture item: %+v", item)
		}
	}
}

func writeProjectsFixture(t *testing.T, data payload.ProjectsResponse) string {
	t.Helper()

	dir := t.TempDir()
	file := filepath.Join(dir, "projects.json")

	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	if err := os.WriteFile(file, body, 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	return file
}
