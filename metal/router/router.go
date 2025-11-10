package router

import (
	"net/http"
	"strings"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/handler"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
)

type Router struct {
	WebsiteRoutes *WebsiteRoutes
	Env           *env.Environment
	Validator     *portal.Validator
	Mux           *http.ServeMux
	Pipeline      middleware.Pipeline
	Db            *database.Connection
}

func (r *Router) PublicPipelineFor(apiHandler endpoint.ApiHandler) http.HandlerFunc {
	return endpoint.NewApiHandler(
		r.Pipeline.Chain(
			apiHandler,
			r.Pipeline.PublicMiddleware.Handle,
		),
	)
}

func (r *Router) PipelineFor(apiHandler endpoint.ApiHandler) http.HandlerFunc {
	tokenMiddleware := middleware.NewTokenMiddleware(
		r.Pipeline.TokenHandler,
		r.Pipeline.ApiKeys,
	)

	return endpoint.NewApiHandler(
		r.Pipeline.Chain(
			apiHandler,
			tokenMiddleware.Handle,
		),
	)
}

func (r *Router) Posts() {
	repo := repository.Posts{DB: r.Db}
	abstract := handler.NewPostsHandler(&repo)

	index := r.PipelineFor(abstract.Index)
	show := r.PipelineFor(abstract.Show)

	r.Mux.HandleFunc("POST /posts", index)
	r.Mux.HandleFunc("GET /posts/{slug}", show)
}

func (r *Router) Categories() {
	repo := repository.Categories{DB: r.Db}
	abstract := handler.NewCategoriesHandler(&repo)

	index := r.PipelineFor(abstract.Index)

	r.Mux.HandleFunc("GET /categories", index)
}

func (r *Router) Signature() {
	abstract := handler.NewSignaturesHandler(r.Validator, r.Pipeline.ApiKeys)
	generate := r.PublicPipelineFor(abstract.Generate)

	r.Mux.HandleFunc("POST /generate-signature", generate)
}

func (r *Router) KeepAlive() {
	abstract := handler.NewKeepAliveHandler(&r.Env.Ping)

	apiHandler := endpoint.NewApiHandler(
		r.Pipeline.Chain(abstract.Handle),
	)

	r.Mux.HandleFunc("GET /ping", apiHandler)
}

func (r *Router) KeepAliveDB() {
	abstract := handler.NewKeepAliveDBHandler(&r.Env.Ping, r.Db)

	apiHandler := endpoint.NewApiHandler(
		r.Pipeline.Chain(abstract.Handle),
	)

	r.Mux.HandleFunc("GET /ping-db", apiHandler)
}

func (r *Router) Metrics() {
	abstract := handler.NewMetricsHandler()

	// Metrics endpoint bypasses middleware - it's for Prometheus/monitoring tools
	apiHandler := endpoint.NewApiHandler(abstract.Handle)

	r.Mux.HandleFunc("GET /metrics", apiHandler)
}

func (r *Router) Profile() {
	maker := handler.NewProfileHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetProfile(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Experience() {
	maker := handler.NewExperienceHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetExperience(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Projects() {
	maker := handler.NewProjectsHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetProjects(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Social() {
	maker := handler.NewSocialHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetSocial(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Talks() {
	maker := handler.NewTalksHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetTalks(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Education() {
	maker := handler.NewEducationHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetEducation(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Recommendations() {
	maker := handler.NewRecommendationsHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetRecommendations(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) composeFixtures(fxt *Fixture, maker func(file string) StaticRouteResource) {
	file := fxt.file
	fullPath := fxt.fullPath

	addStaticRoute(r, file, fullPath, maker)
}

func addStaticRoute[H StaticRouteResource](r *Router, route, fixture string, maker func(string) H) {
	abstract := maker(fixture)
	resolver := r.PipelineFor(abstract.Handle)

	route = strings.TrimLeft(route, "/")
	r.Mux.HandleFunc("GET /"+route, resolver)
}
