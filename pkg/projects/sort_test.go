package projects

import (
	"slices"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestSortByPublishedAtDesc(t *testing.T) {
	projects := []payload.ProjectsData{
		{UUID: "first", PublishedAt: "2026-03-10T00:00:00Z"},
		{UUID: "second", PublishedAt: "2026-03-12T00:00:00Z"},
		{UUID: "third", PublishedAt: ""},
		{UUID: "fourth", PublishedAt: "2026-03-11T00:00:00Z"},
	}

	SortByPublishedAtDesc(projects)

	got := []string{
		projects[0].UUID,
		projects[1].UUID,
		projects[2].UUID,
		projects[3].UUID,
	}

	want := []string{"second", "fourth", "first", "third"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}
