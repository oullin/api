package seo

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"sync"

	"github.com/oullin/handler"
	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/router"
)

type Client struct {
	WebsiteRoutes *router.WebsiteRoutes
	Fixture       router.Fixture
	data          ClientData
}

type ClientData struct {
	profileOnce sync.Once
	profile     *payload.ProfileResponse
	profileErr  error

	projectsOnce sync.Once
	projects     *payload.ProjectsResponse
	projectsErr  error

	recommendationsOnce sync.Once
	recommendations     *payload.RecommendationsResponse
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

func (c *Client) GetTalks() (*payload.TalksResponse, error) {
	return get[payload.TalksResponse](func() router.StaticRouteResource {
		return handler.NewTalksHandler(c.Fixture.GetTalksFile())
	}, "talks")
}

func (c *Client) GetProfile() (*payload.ProfileResponse, error) {
	c.data.profileOnce.Do(func() {
		c.data.profile, c.data.profileErr = get[payload.ProfileResponse](func() router.StaticRouteResource {
			return handler.NewProfileHandler(c.Fixture.GetProfileFile())
		}, "profile")
	})

	return c.data.profile, c.data.profileErr
}

func (c *Client) GetProjects() (*payload.ProjectsResponse, error) {
	c.data.projectsOnce.Do(func() {
		c.data.projects, c.data.projectsErr = get[payload.ProjectsResponse](func() router.StaticRouteResource {
			return handler.NewProjectsHandler(c.Fixture.GetProjectsFile())
		}, "projects")
	})

	return c.data.projects, c.data.projectsErr
}

func (c *Client) GetSocial() (*payload.SocialResponse, error) {
	return get[payload.SocialResponse](func() router.StaticRouteResource {
		return handler.NewSocialHandler(c.Fixture.GetSocialFile())
	}, "social")
}

func (c *Client) GetRecommendations() (*payload.RecommendationsResponse, error) {
	c.data.recommendationsOnce.Do(func() {
		c.data.recommendations, c.data.recommendationsErr = get[payload.RecommendationsResponse](func() router.StaticRouteResource {
			return handler.NewRecommendationsHandler(c.Fixture.GetRecommendationsFile())
		}, "recommendations")
	})

	return c.data.recommendations, c.data.recommendationsErr
}

func (c *Client) GetExperience() (*payload.ExperienceResponse, error) {
	return get[payload.ExperienceResponse](func() router.StaticRouteResource {
		return handler.NewExperienceHandler(c.Fixture.GetExperienceFile())
	}, "experience")
}

func (c *Client) GetEducation() (*payload.EducationResponse, error) {
	return get[payload.EducationResponse](func() router.StaticRouteResource {
		return handler.NewEducationHandler(c.Fixture.GetEducationFile())
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
