package seo

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sync"

	"github.com/oullin/internal/app/router"
	"github.com/oullin/internal/education"
	"github.com/oullin/internal/experience"
	"github.com/oullin/internal/links"
	"github.com/oullin/internal/profile"
	"github.com/oullin/internal/projects"
	"github.com/oullin/internal/recommendations"
	"github.com/oullin/internal/talks"
)

type Client struct {
	WebsiteRoutes *router.WebsiteRoutes
	Fixture       router.Fixture
	data          ClientData
}

type ClientData struct {
	profileOnce sync.Once
	profile     *profile.ProfileResponse
	profileErr  error

	projectsOnce sync.Once
	projects     *projects.ProjectsResponse
	projectsErr  error

	recommendationsOnce sync.Once
	recommendations     *recommendations.RecommendationsResponse
	recommendationsErr  error
}

func NewClient(routes *router.WebsiteRoutes) *Client {
	return &Client{
		WebsiteRoutes: routes,
		Fixture:       routes.Fixture,
	}
}

func get[T any](handler func() router.StaticRouteResource, entityName string) (*T, error) {
	var response T

	if err := fetch[T](&response, handler); err != nil {
		return nil, fmt.Errorf("error fetching %s: %w", entityName, err)
	}

	return &response, nil
}

func (c *Client) GetTalks() (*talks.TalksResponse, error) {
	return get[talks.TalksResponse](func() router.StaticRouteResource {
		return talks.NewTalksHandler(c.Fixture.GetTalksFile())
	}, "talks")
}

func (c *Client) GetProfile() (*profile.ProfileResponse, error) {
	c.data.profileOnce.Do(func() {
		c.data.profile, c.data.profileErr = get[profile.ProfileResponse](func() router.StaticRouteResource {
			return profile.NewProfileHandler(c.Fixture.GetProfileFile())
		}, "profile")
	})

	return c.data.profile, c.data.profileErr
}

func (c *Client) GetProjects() (*projects.ProjectsResponse, error) {
	c.data.projectsOnce.Do(func() {
		c.data.projects, c.data.projectsErr = get[projects.ProjectsResponse](func() router.StaticRouteResource {
			return projects.NewProjectsHandler(c.Fixture.GetProjectsFile())
		}, "projects")
	})

	return c.data.projects, c.data.projectsErr
}

func (c *Client) GetLinks() (*links.LinksResponse, error) {
	return get[links.LinksResponse](func() router.StaticRouteResource {
		return links.NewLinksHandler(c.Fixture.GetLinksFile())
	}, "links")
}

func (c *Client) GetRecommendations() (*recommendations.RecommendationsResponse, error) {
	c.data.recommendationsOnce.Do(func() {
		c.data.recommendations, c.data.recommendationsErr = get[recommendations.RecommendationsResponse](func() router.StaticRouteResource {
			return recommendations.NewRecommendationsHandler(c.Fixture.GetRecommendationsFile())
		}, "recommendations")
	})

	return c.data.recommendations, c.data.recommendationsErr
}

func (c *Client) GetExperience() (*experience.ExperienceResponse, error) {
	return get[experience.ExperienceResponse](func() router.StaticRouteResource {
		return experience.NewExperienceHandler(c.Fixture.GetExperienceFile())
	}, "experience")
}

func (c *Client) GetEducation() (*education.EducationResponse, error) {
	return get[education.EducationResponse](func() router.StaticRouteResource {
		return education.NewEducationHandler(c.Fixture.GetEducationFile())
	}, "education")
}

func fetch[T any](response *T, handler func() router.StaticRouteResource) error {
	req := httptest.NewRequest("GET", "http://localhost:8080/proxy", nil)
	rr := httptest.NewRecorder()

	maker := handler()

	if err := maker.Handle(rr, req); err != nil {
		return err
	}

	if err := json.Unmarshal(rr.Body.Bytes(), response); err != nil {
		return err
	}

	return nil
}
