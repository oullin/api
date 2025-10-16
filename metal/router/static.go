package router

import (
	"net/http"

	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/endpoint"
)

type StaticRouteResource interface {
	Handle(http.ResponseWriter, *http.Request) *endpoint.ApiError
}

type WebsiteRoutes struct {
	OutputDir string
	Lang      string
	SiteName  string
	SiteURL   string
	Fixture   Fixture
}

func NewWebsiteRoutes(e *env.Environment) *WebsiteRoutes {
	return &WebsiteRoutes{
		SiteURL:   e.App.URL,
		SiteName:  e.App.Name,
		Lang:      e.App.Lang(),
		OutputDir: e.Seo.SpaDir,
		Fixture:   NewFixture(),
	}
}
