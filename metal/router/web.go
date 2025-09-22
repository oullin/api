package router

import (
	baseHttp "net/http"

	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/http"
)

const HomePage = "/" // projects, talks, profile
const AboutPage = "about"
const ResumePage = "resume"
const ProjectsPage = "projects"

type StaticRouteResource interface {
	Handle(baseHttp.ResponseWriter, *baseHttp.Request) *http.ApiError
}

type ApiResource struct {
	Path  string
	File  string
	Maker func(string) StaticRouteResource
}

type WebPage struct {
	Path        string
	File        string
	ApiResource map[string]ApiResource
}

type WebsiteRoutes struct {
	OutputDir string
	Lang      string
	SiteName  string
	SiteURL   string
	Fixture   Fixture
	Pages     map[string]WebPage
}

func NewWebsiteRoutes(e *env.Environment) *WebsiteRoutes {
	return &WebsiteRoutes{
		SiteURL:   e.App.URL,
		SiteName:  e.App.Name,
		Lang:      e.App.Lang(),
		OutputDir: e.Seo.SpaDir,
		Fixture:   NewFixture(),
		Pages:     make(map[string]WebPage),
	}
}

func (r *WebsiteRoutes) AddHome() {
	//@todo:  projects, talks, profile

	// 1 - Add the Home page.
	// 2 - Add the Project fixture.
	// 3 - Add the Talks fixture.
	// 4 - Add the Profile fixture.
}

func (r *WebsiteRoutes) AddPageFrom(path, file string, abstract func(string) StaticRouteResource) {
	resource := make(map[string]ApiResource, 1)

	resource[path] = ApiResource{
		Path:  path,
		File:  file,
		Maker: abstract,
	}

	page := WebPage{
		Path:        path,
		File:        file,
		ApiResource: resource,
	}

	r.MapResource(page, resource)
}

func (r *WebsiteRoutes) MapResource(page WebPage, item map[string]ApiResource) {
	//WIP
}
