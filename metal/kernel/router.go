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

type StaticRouteDefinition struct {
	Path  string
	File  string
	Maker func(string) StaticRouteResource
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
		},
		{
			Path: "/experience",
			File: "./storage/fixture/experience.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeExperienceHandler(file)
				return handler
			},
		},
		{
			Path: "/projects",
			File: "./storage/fixture/projects.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeProjectsHandler(file)
				return handler
			},
		},
		{
			Path: "/social",
			File: "./storage/fixture/social.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeSocialHandler(file)
				return handler
			},
		},
		{
			Path: "/talks",
			File: "./storage/fixture/talks.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeTalksHandler(file)
				return handler
			},
		},
		{
			Path: "/education",
			File: "./storage/fixture/education.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeEducationHandler(file)
				return handler
			},
		},
		{
			Path: "/recommendations",
			File: "./storage/fixture/recommendations.json",
			Maker: func(file string) StaticRouteResource {
				handler := handler.MakeRecommendationsHandler(file)
				return handler
			},
		},
	}
}
