package projects

import (
	"testing"

	"github.com/oullin/handler/payload"
)

func TestEnrichResponse_FallsBackToUpdatedAt(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:      "project-1",
				UpdatedAt: "2026-03-10",
			},
		},
	}

	EnrichResponse(response)

	if response.Data[0].PublishedAt != "2026-03-10" {
		t.Fatalf("expected published_at to fall back to updated_at, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_PreservesExplicitPublishedAt(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "project-1",
				PublishedAt: "2026-03-15T00:00:00Z",
				UpdatedAt:   "2026-03-10",
			},
		},
	}

	EnrichResponse(response)

	if response.Data[0].PublishedAt != "2026-03-15T00:00:00Z" {
		t.Fatalf("expected published_at to remain unchanged, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_FallsBackToCreatedAt(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:      "project-1",
				CreatedAt: "2026-03-01",
			},
		},
	}

	EnrichResponse(response)

	if response.Data[0].PublishedAt != "2026-03-01" {
		t.Fatalf("expected published_at to fall back to created_at, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_PrefersUpdatedAtOverCreatedAt(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:      "project-1",
				UpdatedAt: " 2026-03-10 ",
				CreatedAt: " 2026-03-01 ",
			},
		},
	}

	EnrichResponse(response)

	if response.Data[0].PublishedAt != "2026-03-10" {
		t.Fatalf("expected published_at to prefer updated_at, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_NoFallbackWhenAllDatesEmpty(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID: "project-1",
			},
		},
	}

	EnrichResponse(response)

	if response.Data[0].PublishedAt != "" {
		t.Fatalf("expected published_at to remain empty, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_NilResponse(t *testing.T) {
	// The nil input should be ignored without panicking.
	EnrichResponse(nil)
}
