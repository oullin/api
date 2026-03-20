package projects

import (
	"net/url"
	"strings"

	"github.com/oullin/internal/shared/portal"
)

type GitHubRepository struct {
	Owner string
	Name  string
}

func ParseGitHubRepository(rawURL string) (GitHubRepository, bool) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return GitHubRepository{}, false
	}

	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return GitHubRepository{}, false
	}

	switch strings.ToLower(parsed.Hostname()) {
	case "github.com", "www.github.com":
	default:
		return GitHubRepository{}, false
	}

	parts := portal.FilterNonEmpty(strings.Split(parsed.Path, "/"))
	if len(parts) < 2 {
		return GitHubRepository{}, false
	}

	name := strings.TrimSpace(strings.TrimSuffix(parts[1], ".git"))
	owner := strings.TrimSpace(parts[0])

	if owner == "" || name == "" {
		return GitHubRepository{}, false
	}

	return GitHubRepository{
		Owner: owner,
		Name:  name,
	}, true
}
