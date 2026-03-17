package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler"
	"github.com/oullin/handler/payload"
)

func TestProjectsHandler_SortsAndPaginates(t *testing.T) {
	fixture := writeProjectsFixture(t, payload.ProjectsResponse{
		Version: "1.0.0",
		Data: []payload.ProjectsData{
			{UUID: "project-1", Title: "One", URL: "https://github.com/example/one"},
			{UUID: "project-2", Title: "Two", URL: "https://github.com/example/two"},
			{UUID: "project-3", Title: "Three", URL: "https://github.com/example/three"},
			{UUID: "project-4", Title: "Four", URL: "https://github.com/example/four"},
			{UUID: "project-5", Title: "Five", URL: "https://github.com/example/five"},
			{UUID: "project-6", Title: "Six", URL: "https://github.com/example/six"},
			{UUID: "project-7", Title: "Seven", URL: "https://github.com/example/seven"},
			{UUID: "project-8", Title: "Eight", URL: "https://github.com/example/eight"},
			{UUID: "project-9", Title: "Nine", URL: "https://github.com/example/nine"},
			{UUID: "project-10", Title: "Ten", URL: "https://github.com/example/ten"},
		},
	})

	publishedAt := map[string]string{
		"project-1":  "2026-03-01T00:00:00Z",
		"project-2":  "2026-03-10T00:00:00Z",
		"project-3":  "2026-03-03T00:00:00Z",
		"project-4":  "2026-03-08T00:00:00Z",
		"project-5":  "2026-03-02T00:00:00Z",
		"project-6":  "2026-03-07T00:00:00Z",
		"project-7":  "2026-03-06T00:00:00Z",
		"project-8":  "2026-03-09T00:00:00Z",
		"project-9":  "2026-03-05T00:00:00Z",
		"project-10": "2026-03-04T00:00:00Z",
	}

	h := handler.NewProjectsHandlerWithResolver(
		fixture,
		true,
		func(_ context.Context, project payload.ProjectsData) (string, error) {
			return publishedAt[project.UUID], nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/projects?page=1", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	var resp pagination.Pagination[payload.ProjectsData]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
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

	if resp.Data[0].UUID != "project-2" || resp.Data[0].PublishedAt != "2026-03-10T00:00:00Z" {
		t.Fatalf("expected most recent project first, got %+v", resp.Data[0])
	}

	if resp.Data[1].UUID != "project-8" {
		t.Fatalf("expected second most recent project second, got %+v", resp.Data[1])
	}
}

func TestProjectsHandler_Page2(t *testing.T) {
	fixture := writeProjectsFixture(t, payload.ProjectsResponse{
		Version: "1.0.0",
		Data: []payload.ProjectsData{
			{UUID: "project-1", Title: "One", URL: "https://github.com/example/one"},
			{UUID: "project-2", Title: "Two", URL: "https://github.com/example/two"},
			{UUID: "project-3", Title: "Three", URL: "https://github.com/example/three"},
			{UUID: "project-4", Title: "Four", URL: "https://github.com/example/four"},
			{UUID: "project-5", Title: "Five", URL: "https://github.com/example/five"},
			{UUID: "project-6", Title: "Six", URL: "https://github.com/example/six"},
			{UUID: "project-7", Title: "Seven", URL: "https://github.com/example/seven"},
			{UUID: "project-8", Title: "Eight", URL: "https://github.com/example/eight"},
			{UUID: "project-9", Title: "Nine", URL: "https://github.com/example/nine"},
		},
	})

	h := handler.NewProjectsHandlerWithResolver(
		fixture,
		true,
		func(_ context.Context, project payload.ProjectsData) (string, error) {
			return "2026-03-17T12:00:00Z", nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/projects?page=2", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	var resp pagination.Pagination[payload.ProjectsData]
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

func TestProjectsHandler_SortsByPublishedAtWithFallbackDates(t *testing.T) {
	fixture := writeProjectsFixture(t, payload.ProjectsResponse{
		Version: "1.0.0",
		Data: []payload.ProjectsData{
			{
				UUID:      "project-created",
				Title:     "Created",
				URL:       "https://github.com/example/created",
				CreatedAt: "2026-03-01",
			},
			{
				UUID:        "project-updated",
				Title:       "Updated",
				URL:         "https://github.com/example/updated",
				UpdatedAt:   "2026-03-10",
				PublishedAt: "",
			},
			{
				UUID:        "project-published",
				Title:       "Published",
				URL:         "https://github.com/example/published",
				PublishedAt: "2026-03-17T12:00:00Z",
				UpdatedAt:   "2026-03-05",
			},
		},
	})

	h := handler.NewProjectsHandlerWithResolver(
		fixture,
		true,
		func(_ context.Context, project payload.ProjectsData) (string, error) {
			return project.PublishedAt, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/projects", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle: %v", err)
	}

	var resp pagination.Pagination[payload.ProjectsData]
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 items, got %d", len(resp.Data))
	}

	if resp.Data[0].UUID != "project-published" {
		t.Fatalf("expected published project first, got %+v", resp.Data[0])
	}

	if resp.Data[1].UUID != "project-updated" {
		t.Fatalf("expected updated project second, got %+v", resp.Data[1])
	}

	if resp.Data[2].UUID != "project-created" {
		t.Fatalf("expected created project last, got %+v", resp.Data[2])
	}
}

func TestProjectsHandler_NoStoreWhenCacheDisabled(t *testing.T) {
	fixture := writeProjectsFixture(t, payload.ProjectsResponse{
		Version: "1.0.0",
		Data: []payload.ProjectsData{
			{UUID: "project-1", Title: "One", URL: "https://github.com/example/one"},
		},
	})

	h := handler.NewProjectsHandlerWithResolver(
		fixture,
		false,
		func(_ context.Context, project payload.ProjectsData) (string, error) {
			return "2026-03-17T12:00:00Z", nil
		},
	)
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
