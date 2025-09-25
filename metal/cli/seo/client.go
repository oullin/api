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
