package projects

import "testing"

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
			repo, ok := ParseGitHubRepository(tt.rawURL)
			if ok != tt.ok {
				t.Fatalf("expected ok=%t, got %t", tt.ok, ok)
			}

			if !ok {
				return
			}

			if repo.Owner != tt.owner || repo.Name != tt.repo {
				t.Fatalf("unexpected repo: %+v", repo)
			}
		})
	}
}
