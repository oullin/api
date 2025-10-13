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
		return handler.MakeTalksHandler(c.Fixture.GetTalksFile())
	}, "talks")
}

func (c *Client) GetProfile() (*payload.ProfileResponse, error) {
	c.profileOnce.Do(func() {
		c.profile, c.profileErr = get[payload.ProfileResponse](func() router.StaticRouteResource {
			return handler.MakeProfileHandler(c.Fixture.GetProfileFile())
		}, "profile")
	})

	return c.profile, c.profileErr
}

func (c *Client) GetProjects() (*payload.ProjectsResponse, error) {
	c.projectsOnce.Do(func() {
		c.projects, c.projectsErr = get[payload.ProjectsResponse](func() router.StaticRouteResource {
			return handler.MakeProjectsHandler(c.Fixture.GetProjectsFile())
		}, "projects")
	})

	return c.projects, c.projectsErr
}

func (c *Client) GetSocial() (*payload.SocialResponse, error) {
	return get[payload.SocialResponse](func() router.StaticRouteResource {
		return handler.MakeSocialHandler(c.Fixture.GetSocialFile())
	}, "social")
}

func (c *Client) GetRecommendations() (*payload.RecommendationsResponse, error) {
	c.recommendationsOnce.Do(func() {
		c.recommendations, c.recommendationsErr = get[payload.RecommendationsResponse](func() router.StaticRouteResource {
			return handler.MakeRecommendationsHandler(c.Fixture.GetRecommendationsFile())
		}, "recommendations")
	})

	return c.recommendations, c.recommendationsErr
}

func (c *Client) GetExperience() (*payload.ExperienceResponse, error) {
	return get[payload.ExperienceResponse](func() router.StaticRouteResource {
		return handler.MakeExperienceHandler(c.Fixture.GetExperienceFile())
	}, "experience")
}

func (c *Client) GetEducation() (*payload.EducationResponse, error) {
	return get[payload.EducationResponse](func() router.StaticRouteResource {
		return handler.MakeEducationHandler(c.Fixture.GetEducationFile())
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
