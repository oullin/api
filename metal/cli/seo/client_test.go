package seo

import (
	"net/http"
	"strings"
	"testing"

	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/router"
	basehttp "github.com/oullin/pkg/http"
)

type failingRoute struct{ err *basehttp.ApiError }

type invalidJSONRoute struct{ body string }

func (f failingRoute) Handle(http.ResponseWriter, *http.Request) *basehttp.ApiError {
	return f.err
}

func (i invalidJSONRoute) Handle(w http.ResponseWriter, r *http.Request) *basehttp.ApiError {
	_, _ = w.Write([]byte(i.body))

	return nil
}

func TestFetchPropagatesHandlerErrors(t *testing.T) {
	err := fetch[payload.ProfileResponse](
		&payload.ProfileResponse{},
		func() router.StaticRouteResource {
			return failingRoute{err: basehttp.InternalError("boom")}
		},
	)

	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected fetch to return handler error, got %v", err)
	}
}

func TestFetchReturnsJSONDecodeError(t *testing.T) {
	err := fetch[payload.ProfileResponse](
		&payload.ProfileResponse{},
		func() router.StaticRouteResource {
			return invalidJSONRoute{body: "{"}
		},
	)

	if err == nil {
		t.Fatalf("expected json error")
	}
}

func TestClientLoadsFixtures(t *testing.T) {
	withRepoRoot(t)

	e := &env.Environment{
		App: env.AppEnvironment{
			Name:      "SEO Test Fixtures",
			URL:       "https://seo.example.test",
			Type:      "local",
			MasterKey: strings.Repeat("k", 32),
		},
		Seo: env.SeoEnvironment{SpaDir: t.TempDir()},
	}

	routes := router.NewWebsiteRoutes(e)
	client := NewClient(routes)

	profile, err := client.GetProfile()
	if err != nil {
		t.Fatalf("profile err: %v", err)
	}

	if profile.Data.Name == "" || profile.Data.Name != "Gustavo Ocanto" {
		t.Fatalf("unexpected profile data: %+v", profile.Data)
	}

	talks, err := client.GetTalks()
	if err != nil {
		t.Fatalf("talks err: %v", err)
	}

	if len(talks.Data) == 0 {
		t.Fatalf("expected talks data")
	}

	projects, err := client.GetProjects()
	if err != nil {
		t.Fatalf("projects err: %v", err)
	}

	if len(projects.Data) == 0 {
		t.Fatalf("expected projects data")
	}
}
