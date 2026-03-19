package projects

import (
	"strings"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestEnrichResponse_TrimsPublishedAt(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "project-1",
				Sort:        intPtr(1),
				PublishedAt: " 2026-03-10 ",
			},
		},
	}

	if err := EnrichResponse(response); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.Data[0].PublishedAt != "2026-03-10" {
		t.Fatalf("expected published_at to be trimmed, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_SortsBySortAscending(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "project-3",
				Sort:        intPtr(3),
				PublishedAt: "2026-03-15T00:00:00Z",
			},
			{
				UUID:        "project-1",
				Sort:        intPtr(1),
				PublishedAt: "2026-03-01T00:00:00Z",
			},
			{
				UUID:        "project-2",
				Sort:        intPtr(2),
				PublishedAt: "2026-03-10T00:00:00Z",
			},
		},
	}

	if err := EnrichResponse(response); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.Data[0].UUID != "project-1" || response.Data[1].UUID != "project-2" || response.Data[2].UUID != "project-3" {
		t.Fatalf("unexpected sort order: %+v", response.Data)
	}
}

func TestEnrichResponse_UsesPublishedAtTieBreaker(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "older",
				Sort:        intPtr(1),
				PublishedAt: "2026-03-01T00:00:00Z",
			},
			{
				UUID:        "newer",
				Sort:        intPtr(1),
				PublishedAt: "2026-03-10T00:00:00Z",
			},
		},
	}

	if err := EnrichResponse(response); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.Data[0].UUID != "newer" || response.Data[1].UUID != "older" {
		t.Fatalf("expected published_at tie-breaker order, got %+v", response.Data)
	}
}

func TestEnrichResponse_NilResponse(t *testing.T) {
	if err := EnrichResponse(nil); err != nil {
		t.Fatalf("unexpected error for nil response: %v", err)
	}
}

func TestEnrichResponse_RejectsNilSort(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{UUID: "bad-project", Sort: nil, PublishedAt: "2026-03-10"},
		},
	}

	err := EnrichResponse(response)
	if err == nil {
		t.Fatalf("expected error for nil sort")
	}

	if !strings.Contains(err.Error(), "no sort value") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEnrichResponse_RejectsZeroSort(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{UUID: "zero-sort", Sort: intPtr(0), PublishedAt: "2026-03-10"},
		},
	}

	err := EnrichResponse(response)
	if err == nil {
		t.Fatalf("expected error for zero sort")
	}

	if !strings.Contains(err.Error(), "invalid sort value") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEnrichResponse_RejectsEmptyPublishedAt(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{UUID: "no-date", Sort: intPtr(1), PublishedAt: ""},
		},
	}

	err := EnrichResponse(response)
	if err == nil {
		t.Fatalf("expected error for empty published_at")
	}

	if !strings.Contains(err.Error(), "empty published_at") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEnrichResponse_RejectsWhitespaceOnlyPublishedAt(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{UUID: "space-date", Sort: intPtr(1), PublishedAt: "   "},
		},
	}

	err := EnrichResponse(response)
	if err == nil {
		t.Fatalf("expected error for whitespace-only published_at")
	}

	if !strings.Contains(err.Error(), "empty published_at") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
