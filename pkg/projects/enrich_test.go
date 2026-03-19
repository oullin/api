package projects

import (
	"testing"

	"github.com/oullin/handler/payload"
)

func TestEnrichResponse_TrimsPublishedAt(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "project-1",
				PublishedAt: " 2026-03-10 ",
			},
		},
	}

	EnrichResponse(response)

	if response.Data[0].PublishedAt != "2026-03-10" {
		t.Fatalf("expected published_at to be trimmed, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_SortsBySortAscending(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "project-3",
				Sort:        3,
				PublishedAt: "2026-03-15T00:00:00Z",
			},
			{
				UUID:        "project-1",
				Sort:        1,
				PublishedAt: "2026-03-01T00:00:00Z",
			},
			{
				UUID:        "project-2",
				Sort:        2,
				PublishedAt: "2026-03-10T00:00:00Z",
			},
		},
	}

	EnrichResponse(response)

	if response.Data[0].UUID != "project-1" || response.Data[1].UUID != "project-2" || response.Data[2].UUID != "project-3" {
		t.Fatalf("unexpected sort order: %+v", response.Data)
	}
}

func TestEnrichResponse_UsesPublishedAtTieBreaker(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "older",
				Sort:        1,
				PublishedAt: "2026-03-01T00:00:00Z",
			},
			{
				UUID:        "newer",
				Sort:        1,
				PublishedAt: "2026-03-10T00:00:00Z",
			},
		},
	}

	EnrichResponse(response)

	if response.Data[0].UUID != "newer" || response.Data[1].UUID != "older" {
		t.Fatalf("expected published_at tie-breaker order, got %+v", response.Data)
	}
}

func TestEnrichResponse_NilResponse(t *testing.T) {
	// The nil input should be ignored without panicking.
	EnrichResponse(nil)
}
