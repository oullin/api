package web

import (
	baseHttp "net/http"

	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/http"
)

const HomePage = "/"
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

type Page struct {
	Path        string
	File        string
	ApiResource map[string]ApiResource
}

type Routes struct {
	OutputDir string
	Lang      string
	SiteName  string
	SiteURL   string
	Pages     map[string]Page
}

func NewRoutes(e *env.Environment) *Routes {
	return &Routes{
		SiteURL:   e.App.URL,
		SiteName:  e.App.Name,
		Lang:      e.App.Lang(),
		OutputDir: e.Seo.SpaDir,
		Pages:     make(map[string]Page),
	}
}

func (r *Routes) AddPageFrom(path, file string, abstract func(string) StaticRouteResource) {
	resource := make(map[string]ApiResource, 1)

	resource[path] = ApiResource{
		Path:  path,
		File:  file,
		Maker: abstract,
	}

	page := Page{
		Path:        path,
		File:        file,
		ApiResource: resource,
	}

	r.MapResource(page, resource)
}

func (r *Routes) MapResource(page Page, item map[string]ApiResource) {
	//WIP
}
