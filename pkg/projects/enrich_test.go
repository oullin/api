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

func TestEnrichResponse_RejectsInvalidData(t *testing.T) {
	cases := []struct {
		name      string
		data      payload.ProjectsData
		wantMatch string
	}{
		{
			name:      "nil sort",
			data:      payload.ProjectsData{UUID: "bad-project", Sort: nil, PublishedAt: "2026-03-10"},
			wantMatch: "no sort value",
		},
		{
			name:      "zero sort",
			data:      payload.ProjectsData{UUID: "zero-sort", Sort: intPtr(0), PublishedAt: "2026-03-10"},
			wantMatch: "invalid sort value",
		},
		{
			name:      "negative sort",
			data:      payload.ProjectsData{UUID: "negative-sort", Sort: intPtr(-1), PublishedAt: "2026-03-10"},
			wantMatch: "invalid sort value",
		},
		{
			name:      "empty published_at",
			data:      payload.ProjectsData{UUID: "no-date", Sort: intPtr(1), PublishedAt: ""},
			wantMatch: "empty published_at",
		},
		{
			name:      "whitespace-only published_at",
			data:      payload.ProjectsData{UUID: "space-date", Sort: intPtr(1), PublishedAt: "   "},
			wantMatch: "empty published_at",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			response := &payload.ProjectsResponse{
				Data: []payload.ProjectsData{tc.data},
			}

			err := EnrichResponse(response)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}

			if !strings.Contains(err.Error(), tc.wantMatch) {
				t.Fatalf("unexpected error message: %v", err)
			}
		})
	}
}
