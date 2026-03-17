package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/portal"
)

const projectsPageSize = 8
const githubProjectsPublishedAtCacheTTL = time.Hour

type ProjectsPublishedAtResolver func(context.Context, payload.ProjectsData) (string, error)

type githubProjectsPublishedAtResolver struct {
	client *portal.Client
	cache  map[string]projectPublishedAtCacheEntry
	now    func() time.Time
	ttl    time.Duration
	mu     sync.Mutex
}

type projectPublishedAtCacheEntry struct {
	publishedAt string
	expiresAt   time.Time
}

type githubRepository struct {
	owner string
	name  string
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

func NewProjectsHandlerWithResolver(filePath string, cacheEnabled bool, resolver ProjectsPublishedAtResolver) ProjectsHandler {
	if resolver == nil {
		resolver = NewGitHubProjectsPublishedAtResolver()
	}

	return ProjectsHandler{
		filePath:            filePath,
		cacheEnabled:        cacheEnabled,
		publishedAtResolver: resolver,
	}
}

func NewGitHubProjectsPublishedAtResolver() ProjectsPublishedAtResolver {
	client := portal.NewDefaultClient(nil)
	client.AbortOnNone2xx = true
	client.OnHeaders = func(req *http.Request) {
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resolver := &githubProjectsPublishedAtResolver{
		client: client,
		cache:  make(map[string]projectPublishedAtCacheEntry),
		now:    time.Now,
		ttl:    githubProjectsPublishedAtCacheTTL,
	}

	return resolver.Resolve
}

func (r *githubProjectsPublishedAtResolver) Resolve(ctx context.Context, project payload.ProjectsData) (string, error) {
	repo, ok := parseGitHubRepository(project.URL)
	if !ok {
		return "", nil
	}

	cacheKey := repo.owner + "/" + repo.name
	if cached, ok := r.fromCache(cacheKey); ok {
		return cached, nil
	}

	body, err := r.client.Get(
		ctx,
		fmt.Sprintf(
			"https://api.github.com/repos/%s/%s/commits?per_page=1",
			url.PathEscape(repo.owner),
			url.PathEscape(repo.name),
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

	if parsed, ok := parseProjectPublishedAt(publishedAt); ok {
		publishedAt = parsed.UTC().Format(time.RFC3339)
	}

	r.storeCache(cacheKey, publishedAt)

	return publishedAt, nil
}

func (r *githubProjectsPublishedAtResolver) fromCache(key string) (string, bool) {
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

func (r *githubProjectsPublishedAtResolver) storeCache(key string, publishedAt string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache[key] = projectPublishedAtCacheEntry{
		publishedAt: publishedAt,
		expiresAt:   r.now().Add(r.ttl),
	}
}

func enrichProjectsResponse(ctx context.Context, response *payload.ProjectsResponse, resolver ProjectsPublishedAtResolver) {
	if response == nil {
		return
	}

	for i := range response.Data {
		if resolver == nil {
			continue
		}

		publishedAt, err := resolver(ctx, response.Data[i])
		if err != nil {
			slog.Warn(
				"Error resolving project published_at",
				"title", response.Data[i].Title,
				"url", response.Data[i].URL,
				"error", err,
			)

			continue
		}

		if strings.TrimSpace(publishedAt) != "" {
			response.Data[i].PublishedAt = strings.TrimSpace(publishedAt)
		}
	}

	sortProjectsByPublishedAtDesc(response.Data)
}

func paginateProjectsResponse(response payload.ProjectsResponse, paginate pagination.Paginate) payload.ProjectsResponse {
	paginate.Limit = projectsPageSize
	paginate.SetNumItems(int64(len(response.Data)))

	data := response.Data
	start := (paginate.Page - 1) * paginate.Limit

	switch {
	case start < 0:
		start = 0
	case start > len(data):
		start = len(data)
	}

	end := start + paginate.Limit
	if end > len(data) {
		end = len(data)
	}

	items := append([]payload.ProjectsData(nil), data[start:end]...)
	meta := pagination.NewPagination(items, paginate)

	response.Data = items
	response.Page = meta.Page
	response.Total = meta.Total
	response.PageSize = meta.PageSize
	response.TotalPages = meta.TotalPages
	response.NextPage = meta.NextPage
	response.PreviousPage = meta.PreviousPage

	return response
}

func sortProjectsByPublishedAtDesc(projects []payload.ProjectsData) {
	sort.SliceStable(projects, func(i, j int) bool {
		left, leftOK := projectSortDate(projects[i])
		right, rightOK := projectSortDate(projects[j])

		switch {
		case leftOK && rightOK:
			return left.After(right)
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return false
		}
	})
}

func projectSortDate(project payload.ProjectsData) (time.Time, bool) {
	for _, candidate := range []string{project.PublishedAt, project.UpdatedAt, project.CreatedAt} {
		if parsed, ok := parseProjectPublishedAt(candidate); ok {
			return parsed, true
		}
	}

	return time.Time{}, false
}

func parseProjectPublishedAt(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}

func parseGitHubRepository(rawURL string) (githubRepository, bool) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return githubRepository{}, false
	}

	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return githubRepository{}, false
	}

	switch strings.ToLower(parsed.Hostname()) {
	case "github.com", "www.github.com":
	default:
		return githubRepository{}, false
	}

	parts := portal.FilterNonEmpty(strings.Split(parsed.Path, "/"))
	if len(parts) < 2 {
		return githubRepository{}, false
	}

	name := strings.TrimSpace(strings.TrimSuffix(parts[1], ".git"))
	owner := strings.TrimSpace(parts[0])
	if owner == "" || name == "" {
		return githubRepository{}, false
	}

	return githubRepository{
		owner: owner,
		name:  name,
	}, true
}
