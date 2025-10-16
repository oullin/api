package seo

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/endpoint"
)

type failingRoute struct{ err *endpoint.ApiError }

type invalidJSONRoute struct{ body string }

func (f failingRoute) Handle(http.ResponseWriter, *http.Request) *endpoint.ApiError {
	return f.err
}

func (i invalidJSONRoute) Handle(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
	_, _ = w.Write([]byte(i.body))

	return nil
}

func TestFetchPropagatesHandlerErrors(t *testing.T) {
	err := fetch[payload.ProfileResponse](
		&payload.ProfileResponse{},
		func() router.StaticRouteResource {
			return failingRoute{err: endpoint.InternalError("boom")}
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

	spaDir := t.TempDir()
	imagesDir := filepath.Join(spaDir, "posts", "images")

	e := &env.Environment{
		App: env.AppEnvironment{
			Name:      "SEO Test Fixtures",
			URL:       "https://seo.example.test",
			Type:      "local",
			MasterKey: strings.Repeat("k", 32),
		},
		Seo: env.SeoEnvironment{SpaDir: spaDir, SpaImagesDir: imagesDir},
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

	social, err := client.GetSocial()
	if err != nil {
		t.Fatalf("social err: %v", err)
	}

	if len(social.Data) == 0 {
		t.Fatalf("expected social data")
	}

	recs, err := client.GetRecommendations()
	if err != nil {
		t.Fatalf("recommendations err: %v", err)
	}

	if len(recs.Data) == 0 {
		t.Fatalf("expected recommendations data")
	}

	experience, err := client.GetExperience()
	if err != nil {
		t.Fatalf("experience err: %v", err)
	}

	if len(experience.Data) == 0 {
		t.Fatalf("expected experience data")
	}

	education, err := client.GetEducation()
	if err != nil {
		t.Fatalf("education err: %v", err)
	}

	if len(education.Data) == 0 {
		t.Fatalf("expected education data")
	}
}
