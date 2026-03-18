package projects

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/portal"
)

func TestGitHubPublishedAtResolverUsesOldestDefaultBranchCommit(t *testing.T) {
	var repoRequests atomic.Int32
	var commitsRequests atomic.Int32

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/example/project":
			repoRequests.Add(1)
			_, _ = w.Write([]byte(`{"default_branch":"main"}`))
		case "/repos/example/project/commits":
			commitsRequests.Add(1)
			if r.URL.Query().Get("sha") != "main" || r.URL.Query().Get("per_page") != "1" {
				t.Fatalf("unexpected commits query %q", r.URL.RawQuery)
			}

			switch r.URL.Query().Get("page") {
			case "":
				w.Header().Set("Link", `<`+srv.URL+`/repos/example/project/commits?sha=main&per_page=1&page=2>; rel="last"`)
				_, _ = w.Write([]byte(`[{"commit":{"committer":{"date":"2026-03-12T00:00:00Z"}}}]`))
			case "2":
				_, _ = w.Write([]byte(`[{"commit":{"committer":{"date":"2024-01-02T03:04:05Z"}}}]`))
			default:
				t.Fatalf("unexpected page %q", r.URL.Query().Get("page"))
			}
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	resolver := &githubPublishedAtResolver{
		client:     portal.NewDefaultClient(nil),
		cache:      make(map[string]cacheEntry),
		now:        func() time.Time { return time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC) },
		ttl:        cacheTTL,
		apiBaseURL: srv.URL,
	}

	got, err := resolver.Resolve(context.Background(), payload.ProjectsData{
		URL: "https://github.com/example/project",
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if got != "2024-01-02T03:04:05Z" {
		t.Fatalf("unexpected published_at %q", got)
	}

	if repoRequests.Load() != 1 {
		t.Fatalf("expected 1 repo request, got %d", repoRequests.Load())
	}

	if commitsRequests.Load() != 2 {
		t.Fatalf("expected 2 commit requests, got %d", commitsRequests.Load())
	}
}

func TestGitHubPublishedAtResolverHandlesSingleCommitRepository(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/example/project":
			_, _ = w.Write([]byte(`{"default_branch":"main"}`))
		case "/repos/example/project/commits":
			_, _ = w.Write([]byte(`[{"commit":{"author":{"date":"2025-05-06T07:08:09Z"}}}]`))
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	resolver := &githubPublishedAtResolver{
		client:     portal.NewDefaultClient(nil),
		cache:      make(map[string]cacheEntry),
		now:        time.Now,
		ttl:        cacheTTL,
		apiBaseURL: srv.URL,
	}

	got, err := resolver.Resolve(context.Background(), payload.ProjectsData{
		URL: "https://github.com/example/project",
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if got != "2025-05-06T07:08:09Z" {
		t.Fatalf("unexpected published_at %q", got)
	}
}

func TestGitHubPublishedAtResolverUsesCache(t *testing.T) {
	var hits atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case "/repos/example/project":
			_, _ = w.Write([]byte(`{"default_branch":"main"}`))
		case "/repos/example/project/commits":
			_, _ = w.Write([]byte(`[{"commit":{"committer":{"date":"2025-01-01T00:00:00Z"}}}]`))
		default:
			t.Fatalf("unexpected request path %q", r.URL.Path)
		}
	}))
	defer srv.Close()

	now := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	resolver := &githubPublishedAtResolver{
		client:     portal.NewDefaultClient(nil),
		cache:      make(map[string]cacheEntry),
		now:        func() time.Time { return now },
		ttl:        cacheTTL,
		apiBaseURL: srv.URL,
	}

	project := payload.ProjectsData{URL: "https://github.com/example/project"}

	first, err := resolver.Resolve(context.Background(), project)
	if err != nil {
		t.Fatalf("first resolve: %v", err)
	}

	second, err := resolver.Resolve(context.Background(), project)
	if err != nil {
		t.Fatalf("second resolve: %v", err)
	}

	if first != second {
		t.Fatalf("expected cached result, got %q then %q", first, second)
	}

	if hits.Load() != 2 {
		t.Fatalf("expected 2 upstream hits, got %d", hits.Load())
	}
}

func TestGitHubPublishedAtResolverSkipsNonGitHubURLs(t *testing.T) {
	resolver := &githubPublishedAtResolver{
		client:     portal.NewDefaultClient(nil),
		cache:      make(map[string]cacheEntry),
		now:        time.Now,
		ttl:        cacheTTL,
		apiBaseURL: "http://127.0.0.1:1",
	}

	got, err := resolver.Resolve(context.Background(), payload.ProjectsData{
		URL: "https://example.com/project",
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if got != "" {
		t.Fatalf("expected empty published_at, got %q", got)
	}
}
