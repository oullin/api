package projects

import (
	"context"
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

	EnrichResponse(context.Background(), response, func(_ context.Context, _ payload.ProjectsData) (string, error) {
		return "", nil
	})

	if response.Data[0].PublishedAt != "2026-03-10" {
		t.Fatalf("expected published_at to fall back to updated_at, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_NoFallbackWhenPublishedAtSet(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "project-1",
				PublishedAt: "2026-03-15T00:00:00Z",
				UpdatedAt:   "2026-03-10",
			},
		},
	}

	EnrichResponse(context.Background(), response, func(_ context.Context, p payload.ProjectsData) (string, error) {
		return p.PublishedAt, nil
	})

	if response.Data[0].PublishedAt != "2026-03-15T00:00:00Z" {
		t.Fatalf("expected published_at to remain unchanged, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_PrefersExplicitPublishedAtOverResolver(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID:        "project-1",
				PublishedAt: "2025-09-15",
				UpdatedAt:   "2026-03-10",
			},
		},
	}

	EnrichResponse(context.Background(), response, func(_ context.Context, _ payload.ProjectsData) (string, error) {
		return "2026-03-17T12:00:00Z", nil
	})

	if response.Data[0].PublishedAt != "2025-09-15" {
		t.Fatalf("expected explicit published_at to remain unchanged, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_NoFallbackWhenUpdatedAtEmpty(t *testing.T) {
	response := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{
			{
				UUID: "project-1",
			},
		},
	}

	EnrichResponse(context.Background(), response, func(_ context.Context, _ payload.ProjectsData) (string, error) {
		return "", nil
	})

	if response.Data[0].PublishedAt != "" {
		t.Fatalf("expected published_at to remain empty, got %q", response.Data[0].PublishedAt)
	}
}

func TestEnrichResponse_NilResponse(t *testing.T) {
	EnrichResponse(context.Background(), nil, func(_ context.Context, _ payload.ProjectsData) (string, error) {
		return "2026-03-10", nil
	})
}
