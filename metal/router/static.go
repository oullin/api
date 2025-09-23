package router

import (
	baseHttp "net/http"

	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/http"
)

type StaticRouteResource interface {
	Handle(baseHttp.ResponseWriter, *baseHttp.Request) *http.ApiError
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
