package projects

import (
	"slices"
	"testing"
)

func intPtrSort(v int) *int { return &v }

func TestSortBySortAsc(t *testing.T) {
	projects := []ProjectsData{
		{UUID: "third", Sort: intPtrSort(3), PublishedAt: "2026-03-10T00:00:00Z"},
		{UUID: "first", Sort: intPtrSort(1), PublishedAt: "2026-03-12T00:00:00Z"},
		{UUID: "fourth", Sort: intPtrSort(4), PublishedAt: ""},
		{UUID: "second", Sort: intPtrSort(2), PublishedAt: "2026-03-11T00:00:00Z"},
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
	projects := []ProjectsData{
		{UUID: "older", Sort: intPtrSort(1), PublishedAt: "2026-03-10T00:00:00Z"},
		{UUID: "newer", Sort: intPtrSort(1), PublishedAt: "2026-03-12T00:00:00Z"},
		{UUID: "other", Sort: intPtrSort(2), PublishedAt: "2026-03-11T00:00:00Z"},
	}

	SortBySortAsc(projects)

	got := []string{projects[0].UUID, projects[1].UUID, projects[2].UUID}
	want := []string{"newer", "older", "other"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortBySortAsc_StableWhenPublishedAtInvalid(t *testing.T) {
	projects := []ProjectsData{
		{UUID: "first", Sort: intPtrSort(1), PublishedAt: "bad"},
		{UUID: "second", Sort: intPtrSort(1), PublishedAt: ""},
	}

	SortBySortAsc(projects)

	got := []string{projects[0].UUID, projects[1].UUID}
	want := []string{"first", "second"}
	if !slices.Equal(got, want) {
		t.Fatalf("unexpected order: got %v want %v", got, want)
	}
}

func TestSortBySortAsc_NilSortSinksToEnd(t *testing.T) {
	projects := []ProjectsData{
		{UUID: "no-sort", Sort: nil, PublishedAt: "2026-03-10T00:00:00Z"},
		{UUID: "has-sort", Sort: intPtrSort(1), PublishedAt: "2026-03-10T00:00:00Z"},
	}

	SortBySortAsc(projects)

	got := []string{projects[0].UUID, projects[1].UUID}
	want := []string{"has-sort", "no-sort"}
	if !slices.Equal(got, want) {
		t.Fatalf("expected nil sort to sink to end: got %v want %v", got, want)
	}
}
