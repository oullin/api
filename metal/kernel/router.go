package kernel

import (
	baseHttp "net/http"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/handler"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
)

type StaticRouteResource interface {
	Handle(baseHttp.ResponseWriter, *baseHttp.Request) *http.ApiError
}

type Hreflang struct {
	Lang string
	Href string
}

type Favicon struct {
	Rel   string
	Href  string
	Type  string
	Sizes string
}

type OGSpec struct {
	Type        string
	Image       string
	ImageAlt    string
	ImageWidth  string
	ImageHeight string
	SiteName    string
	Locale      string
}

type TwitterSpec struct {
	Card     string
	Image    string
	ImageAlt string
}

type SharePage struct {
	Lang           string
	Title          string
	Description    string
	Canonical      string
	Robots         string
	ThemeColor     string
	JsonLD         string
	OG             OGSpec
	Twitter        TwitterSpec
	Hreflangs      []Hreflang
	Favicons       []Favicon
	Manifest       string
	AppleTouchIcon string
	BuildRev       string
}

type StaticRouteDefinition struct {
	Path  string
	File  string
	Maker func(string) StaticRouteResource
	Page  SharePage
}

func addStaticRoute(r *Router, route StaticRouteDefinition) {
	resource := route.Maker(route.File)
	resolver := r.PipelineFor(resource.Handle)
	r.Mux.HandleFunc("GET "+route.Path, resolver)
}

type Router struct {
	Env       *env.Environment
	Mux       *baseHttp.ServeMux
	Pipeline  middleware.Pipeline
	Db        *database.Connection
	validator *portal.Validator
}

func (r *Router) PublicPipelineFor(apiHandler http.ApiHandler) baseHttp.HandlerFunc {
	return http.MakeApiHandler(
		r.Pipeline.Chain(
			apiHandler,
			r.Pipeline.PublicMiddleware.Handle,
		),
	)
}

func (r *Router) PipelineFor(apiHandler http.ApiHandler) baseHttp.HandlerFunc {
	tokenMiddleware := middleware.MakeTokenMiddleware(
		r.Pipeline.TokenHandler,
		r.Pipeline.ApiKeys,
	)

	return http.MakeApiHandler(
		r.Pipeline.Chain(
			apiHandler,
			tokenMiddleware.Handle,
		),
	)
}

func (r *Router) Posts() {
	repo := repository.Posts{DB: r.Db}
	abstract := handler.MakePostsHandler(&repo)

	index := r.PipelineFor(abstract.Index)
	show := r.PipelineFor(abstract.Show)

	r.Mux.HandleFunc("POST /posts", index)
	r.Mux.HandleFunc("GET /posts/{slug}", show)
}

func (r *Router) Categories() {
	repo := repository.Categories{DB: r.Db}
	abstract := handler.MakeCategoriesHandler(&repo)

	index := r.PipelineFor(abstract.Index)

	r.Mux.HandleFunc("GET /categories", index)
}

func (r *Router) Signature() {
	abstract := handler.MakeSignaturesHandler(r.validator, r.Pipeline.ApiKeys)
	generate := r.PublicPipelineFor(abstract.Generate)

	r.Mux.HandleFunc("POST /generate-signature", generate)
}

func (r *Router) KeepAlive() {
	abstract := handler.MakeKeepAliveHandler(&r.Env.Ping)

	apiHandler := http.MakeApiHandler(
		r.Pipeline.Chain(abstract.Handle),
	)

	r.Mux.HandleFunc("GET /ping", apiHandler)
}

func (r *Router) KeepAliveDB() {
	abstract := handler.MakeKeepAliveDBHandler(&r.Env.Ping, r.Db)

	apiHandler := http.MakeApiHandler(
		r.Pipeline.Chain(abstract.Handle),
	)

	r.Mux.HandleFunc("GET /ping-db", apiHandler)
}

func (r *Router) StaticRoutes() {
	for _, route := range StaticRouteDefinitions() {
		addStaticRoute(r, route)
	}
}

func StaticRouteDefinitions() []StaticRouteDefinition {
	return []StaticRouteDefinition{
		{
			Path: "/profile",
			File: "./storage/fixture/profile.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeProfileHandler(file)
				return handler
			},
			Page: SharePage{
				Lang:        "en",
				Title:       "Professional profile",
				Description: "Review Gustavo Ocanto's profile, skills, and contact information.",
			},
		},
		{
			Path: "/experience",
			File: "./storage/fixture/experience.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeExperienceHandler(file)
				return handler
			},
			Page: SharePage{
				Lang:        "en",
				Title:       "Experience & leadership",
				Description: "Explore career experience and leadership milestones from Gustavo Ocanto.",
			},
		},
		{
			Path: "/projects",
			File: "./storage/fixture/projects.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeProjectsHandler(file)
				return handler
			},
			Page: SharePage{
				Lang:        "en",
				Title:       "Highlighted projects",
				Description: "Browse selected projects delivered by Gustavo Ocanto.",
			},
		},
		{
			Path: "/social",
			File: "./storage/fixture/social.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeSocialHandler(file)
				return handler
			},
			Page: SharePage{
				Lang:        "en",
				Title:       "Social presence",
				Description: "Follow Gustavo Ocanto across social channels.",
			},
		},
		{
			Path: "/talks",
			File: "./storage/fixture/talks.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeTalksHandler(file)
				return handler
			},
			Page: SharePage{
				Lang:        "en",
				Title:       "Talks & presentations",
				Description: "Catch recordings and details from Gustavo Ocanto's speaking engagements.",
			},
		},
		{
			Path: "/education",
			File: "./storage/fixture/education.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeEducationHandler(file)
				return handler
			},
			Page: SharePage{
				Lang:        "en",
				Title:       "Education & certifications",
				Description: "See the academic background and certifications earned by Gustavo Ocanto.",
			},
		},
		{
			Path: "/recommendations",
			File: "./storage/fixture/recommendations.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeRecommendationsHandler(file)
				return handler
			},
			Page: SharePage{
				Lang:        "en",
				Title:       "Recommendations",
				Description: "Read testimonials and endorsements for Gustavo Ocanto.",
			},
		},
	}
}
