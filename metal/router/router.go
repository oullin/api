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
	metricsHandler := handler.NewMetricsHandler()

	// Metrics endpoint blocked from public access by Caddy (see @protected matcher in Caddyfile)
	// Only accessible internally via direct container access (api:8080/metrics)
	// Prometheus scrapes via internal DNS without going through Caddy's public listener
	r.Mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, req *http.Request) {
		_ = metricsHandler.Handle(w, req)
	})
}

func (r *Router) Profile() {
	maker := handler.NewProfileHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetProfile(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Experience() {
	maker := handler.NewExperienceHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetExperience(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Projects() {
	maker := handler.NewProjectsHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetProjects(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Links() {
	maker := handler.NewLinksHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetLinks(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Talks() {
	maker := handler.NewTalksHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetTalks(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Education() {
	maker := handler.NewEducationHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetEducation(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Recommendations() {
	maker := handler.NewRecommendationsHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetRecommendations(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) composeFixtures(fxt *Fixture, maker func(file string, cacheEnabled bool) StaticRouteResource) {
	file := fxt.file
	fullPath := fxt.fullPath

	addStaticRoute(r, file, fullPath, maker)
}

func addStaticRoute[H StaticRouteResource](r *Router, route, fixture string, maker func(string, bool) H) {
	abstract := maker(fixture, !r.Env.App.IsLocal())
	resolver := r.PipelineFor(abstract.Handle)

	route = strings.TrimLeft(route, "/")
	r.Mux.HandleFunc("GET /"+route, resolver)
}
