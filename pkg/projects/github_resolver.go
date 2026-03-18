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
const githubAPIBaseURL = "https://api.github.com"

type githubPublishedAtResolver struct {
	client     *portal.Client
	cache      map[string]cacheEntry
	now        func() time.Time
	ttl        time.Duration
	apiBaseURL string
	mu         sync.Mutex
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

type githubRepositoryResponse struct {
	DefaultBranch string `json:"default_branch"`
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
		client:     client,
		cache:      make(map[string]cacheEntry),
		now:        time.Now,
		ttl:        cacheTTL,
		apiBaseURL: githubAPIBaseURL,
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

	repoBody, err := r.client.Get(ctx, r.repositoryURL(repo))

	if err != nil {
		return "", fmt.Errorf("fetch repository metadata for %s: %w", cacheKey, err)
	}

	var repository githubRepositoryResponse
	if err := json.Unmarshal([]byte(repoBody), &repository); err != nil {
		return "", fmt.Errorf("decode repository metadata for %s: %w", cacheKey, err)
	}

	defaultBranch := strings.TrimSpace(repository.DefaultBranch)
	if defaultBranch == "" {
		return "", nil
	}

	commitsResp, err := r.client.GetResponse(ctx, r.commitsURL(repo, defaultBranch))
	if err != nil {
		return "", fmt.Errorf("fetch latest commit page for %s: %w", cacheKey, err)
	}

	commits, err := decodeGitHubCommits(commitsResp.Body)
	if err != nil {
		return "", fmt.Errorf("decode latest commit page for %s: %w", cacheKey, err)
	}

	lastPageURL := findGitHubLink(commitsResp.Header.Get("Link"), "last")
	if lastPageURL != "" {
		lastPageBody, err := r.client.Get(ctx, lastPageURL)
		if err != nil {
			return "", fmt.Errorf("fetch oldest commit page for %s: %w", cacheKey, err)
		}

		commits, err = decodeGitHubCommits(lastPageBody)
		if err != nil {
			return "", fmt.Errorf("decode oldest commit page for %s: %w", cacheKey, err)
		}
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

func (r *githubPublishedAtResolver) repositoryURL(repo GitHubRepository) string {
	return fmt.Sprintf(
		"%s/repos/%s/%s",
		strings.TrimSuffix(r.apiBaseURL, "/"),
		url.PathEscape(repo.Owner),
		url.PathEscape(repo.Name),
	)
}

func (r *githubPublishedAtResolver) commitsURL(repo GitHubRepository, branch string) string {
	values := url.Values{}
	values.Set("sha", branch)
	values.Set("per_page", "1")

	return fmt.Sprintf(
		"%s/repos/%s/%s/commits?%s",
		strings.TrimSuffix(r.apiBaseURL, "/"),
		url.PathEscape(repo.Owner),
		url.PathEscape(repo.Name),
		values.Encode(),
	)
}

func decodeGitHubCommits(body string) ([]githubCommitResponse, error) {
	var commits []githubCommitResponse
	if err := json.Unmarshal([]byte(body), &commits); err != nil {
		return nil, err
	}

	return commits, nil
}

func findGitHubLink(header string, rel string) string {
	for _, part := range strings.Split(header, ",") {
		section := strings.TrimSpace(part)
		if section == "" {
			continue
		}

		pieces := strings.Split(section, ";")
		if len(pieces) < 2 {
			continue
		}

		target := strings.TrimSpace(pieces[0])
		if !strings.HasPrefix(target, "<") || !strings.HasSuffix(target, ">") {
			continue
		}

		for _, piece := range pieces[1:] {
			if strings.TrimSpace(piece) == fmt.Sprintf(`rel="%s"`, rel) {
				return strings.Trim(target, "<>")
			}
		}
	}

	return ""
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
