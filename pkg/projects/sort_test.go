package projects

import (
	"slices"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestSortBySortAsc(t *testing.T) {
	projects := []payload.ProjectsData{
		{UUID: "third", Sort: 3, PublishedAt: "2026-03-10T00:00:00Z"},
		{UUID: "first", Sort: 1, PublishedAt: "2026-03-12T00:00:00Z"},
		{UUID: "fourth", Sort: 4, PublishedAt: ""},
		{UUID: "second", Sort: 2, PublishedAt: "2026-03-11T00:00:00Z"},
	}

	SortBySortAsc(projects)

	got := []string{
		projects[0].UUID,
		projects[1].UUID,
		projects[2].UUID,
		projects[3].UUID,
	}

	want := []string{"first", "second", "third", "fourth"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortBySortAsc_UsesPublishedAtDescAsTieBreaker(t *testing.T) {
	projects := []payload.ProjectsData{
		{UUID: "older", Sort: 1, PublishedAt: "2026-03-10T00:00:00Z"},
		{UUID: "newer", Sort: 1, PublishedAt: "2026-03-12T00:00:00Z"},
		{UUID: "other", Sort: 2, PublishedAt: "2026-03-11T00:00:00Z"},
	}

	SortBySortAsc(projects)

	got := []string{projects[0].UUID, projects[1].UUID, projects[2].UUID}
	want := []string{"newer", "older", "other"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortBySortAsc_StableWhenPublishedAtInvalid(t *testing.T) {
	projects := []payload.ProjectsData{
		{UUID: "first", Sort: 1, PublishedAt: "bad"},
		{UUID: "second", Sort: 1, PublishedAt: ""},
	}

	SortBySortAsc(projects)

	got := []string{projects[0].UUID, projects[1].UUID}
	want := []string{"first", "second"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}
