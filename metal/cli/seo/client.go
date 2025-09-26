package seo

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/oullin/handler"
	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/router"
)

type Client struct {
	WebsiteRoutes *router.WebsiteRoutes
	Fixture       router.Fixture
}

func NewClient(routes *router.WebsiteRoutes) *Client {
	return &Client{
		WebsiteRoutes: routes,
		Fixture:       routes.Fixture,
	}
}

func (c *Client) GetTalks() (*payload.TalksResponse, error) {
	var talks payload.TalksResponse

	fn := func() router.StaticRouteResource {
		return handler.MakeTalksHandler(c.Fixture.GetTalksFile())
	}

	if err := fetch[payload.TalksResponse](&talks, fn); err != nil {
		return nil, fmt.Errorf("home: error fetching talks: %w", err)
	}

	return &talks, nil
}

func (c *Client) GetProfile() (*payload.ProfileResponse, error) {
	var profile payload.ProfileResponse

	fn := func() router.StaticRouteResource {
		return handler.MakeProfileHandler(c.Fixture.GetProfileFile())
	}

	if err := fetch[payload.ProfileResponse](&profile, fn); err != nil {
		return nil, fmt.Errorf("error fetching profile: %w", err)
	}

	return &profile, nil
}

func (c *Client) GetProjects() (*payload.ProjectsResponse, error) {
	var projects payload.ProjectsResponse

	fn := func() router.StaticRouteResource {
		return handler.MakeProjectsHandler(c.Fixture.GetProjectsFile())
	}

	if err := fetch[payload.ProjectsResponse](&projects, fn); err != nil {
		return nil, fmt.Errorf("error fetching projects: %w", err)
	}

	return &projects, nil
}

func (c *Client) GetSocial() (*payload.SocialResponse, error) {
	var social payload.SocialResponse

	fn := func() router.StaticRouteResource {
		return handler.MakeSocialHandler(c.Fixture.GetSocialFile())
	}

	if err := fetch[payload.SocialResponse](&social, fn); err != nil {
		return nil, fmt.Errorf("error fetching social: %w", err)
	}

	return &social, nil
}

func (c *Client) GetRecommendations() (*payload.RecommendationsResponse, error) {
	var recs payload.RecommendationsResponse

	fn := func() router.StaticRouteResource {
		return handler.MakeRecommendationsHandler(c.Fixture.GetRecommendationsFile())
	}

	if err := fetch[payload.RecommendationsResponse](&recs, fn); err != nil {
		return nil, fmt.Errorf("error fetching recommendations: %w", err)
	}

	return &recs, nil
}

func (c *Client) GetExperience() (*payload.ExperienceResponse, error) {
	var exp payload.ExperienceResponse

	fn := func() router.StaticRouteResource {
		return handler.MakeExperienceHandler(c.Fixture.GetExperienceFile())
	}

	if err := fetch[payload.ExperienceResponse](&exp, fn); err != nil {
		return nil, fmt.Errorf("error fetching experience: %w", err)
	}

	return &exp, nil
}

func (c *Client) GetEducation() (*payload.EducationResponse, error) {
	var edu payload.EducationResponse

	fn := func() router.StaticRouteResource {
		return handler.MakeEducationHandler(c.Fixture.GetEducationFile())
	}

	if err := fetch[payload.EducationResponse](&edu, fn); err != nil {
		return nil, fmt.Errorf("error fetching education: %w", err)
	}

	return &edu, nil
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
