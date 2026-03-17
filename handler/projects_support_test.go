package handler

import (
	"testing"

	"github.com/oullin/handler/payload"
)

func TestParseGitHubRepository(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
		owner  string
		repo   string
		ok     bool
	}{
		{
			name:   "repo root",
			rawURL: "https://github.com/oullin/api",
			owner:  "oullin",
			repo:   "api",
			ok:     true,
		},
		{
			name:   "repo sub path",
			rawURL: "https://github.com/laravel/framework/pulls?q=is%3Apr+is%3Aclosed+author%3Agocanto",
			owner:  "laravel",
			repo:   "framework",
			ok:     true,
		},
		{
			name:   "account only",
			rawURL: "https://github.com/aurachakra",
			ok:     false,
		},
		{
			name:   "non github",
			rawURL: "https://example.com/repo",
			ok:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ok := parseGitHubRepository(tt.rawURL)
			if ok != tt.ok {
				t.Fatalf("expected ok=%t, got %t", tt.ok, ok)
			}

			if !ok {
				return
			}

			if repo.owner != tt.owner || repo.name != tt.repo {
				t.Fatalf("unexpected repo: %+v", repo)
			}
		})
	}
}

func TestSortProjectsByPublishedAtDesc(t *testing.T) {
	projects := []payload.ProjectsData{
		{UUID: "first", PublishedAt: "2026-03-10T00:00:00Z"},
		{UUID: "second", PublishedAt: "2026-03-12T00:00:00Z"},
		{UUID: "third", PublishedAt: ""},
		{UUID: "fourth", PublishedAt: "2026-03-11T00:00:00Z"},
	}

	sortProjectsByPublishedAtDesc(projects)

	got := []string{
		projects[0].UUID,
		projects[1].UUID,
		projects[2].UUID,
		projects[3].UUID,
	}

	want := []string{"second", "fourth", "first", "third"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected order: got %v want %v", got, want)
		}
	}
}
