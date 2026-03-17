package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/portal"
)

const cacheTTL = time.Hour

type githubPublishedAtResolver struct {
	client *portal.Client
	cache  map[string]cacheEntry
	now    func() time.Time
	ttl    time.Duration
	mu     sync.Mutex
}

type cacheEntry struct {
	publishedAt string
	expiresAt   time.Time
}

type githubCommitResponse struct {
	Commit struct {
		Author struct {
			Date string `json:"date"`
		} `json:"author"`
		Committer struct {
			Date string `json:"date"`
		} `json:"committer"`
	} `json:"commit"`
}

func NewGitHubPublishedAtResolver() PublishedAtResolver {
	client := portal.NewDefaultClient(nil)
	client.AbortOnNone2xx = true

	client.OnHeaders = func(req *http.Request) {
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resolver := &githubPublishedAtResolver{
		client: client,
		cache:  make(map[string]cacheEntry),
		now:    time.Now,
		ttl:    cacheTTL,
	}

	return resolver.Resolve
}

func (r *githubPublishedAtResolver) Resolve(ctx context.Context, project payload.ProjectsData) (string, error) {
	repo, ok := ParseGitHubRepository(project.URL)

	if !ok {
		return "", nil
	}

	cacheKey := repo.Owner + "/" + repo.Name
	if cached, ok := r.fromCache(cacheKey); ok {
		return cached, nil
	}

	body, err := r.client.Get(
		ctx,
		fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/commits?per_page=1",
			url.PathEscape(repo.Owner),
			url.PathEscape(repo.Name),
		),
	)

	if err != nil {
		return "", fmt.Errorf("fetch latest commit for %s: %w", cacheKey, err)
	}

	var commits []githubCommitResponse
	if err := json.Unmarshal([]byte(body), &commits); err != nil {
		return "", fmt.Errorf("decode latest commit for %s: %w", cacheKey, err)
	}

	if len(commits) == 0 {
		return "", nil
	}

	publishedAt := strings.TrimSpace(commits[0].Commit.Committer.Date)
	if publishedAt == "" {
		publishedAt = strings.TrimSpace(commits[0].Commit.Author.Date)
	}

	if publishedAt == "" {
		return "", nil
	}

	if parsed, ok := ParsePublishedAt(publishedAt); ok {
		publishedAt = parsed.UTC().Format(time.RFC3339)
	}

	r.storeCache(cacheKey, publishedAt)

	return publishedAt, nil
}

func (r *githubPublishedAtResolver) fromCache(key string) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.cache[key]
	if !ok {
		return "", false
	}

	if r.now().After(entry.expiresAt) {
		delete(r.cache, key)

		return "", false
	}

	return entry.publishedAt, true
}

func (r *githubPublishedAtResolver) storeCache(key string, publishedAt string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[key] = cacheEntry{
		publishedAt: publishedAt,
		expiresAt:   r.now().Add(r.ttl),
	}
}
