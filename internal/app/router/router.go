package router

import (
	"net/http"
	"strings"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	env "github.com/oullin/internal/app/config"
	"github.com/oullin/internal/app/middleware"
	"github.com/oullin/internal/categories"
	"github.com/oullin/internal/education"
	"github.com/oullin/internal/experience"
	"github.com/oullin/internal/health"
	"github.com/oullin/internal/links"
	"github.com/oullin/internal/metrics"
	"github.com/oullin/internal/posts"
	"github.com/oullin/internal/profile"
	"github.com/oullin/internal/projects"
	"github.com/oullin/internal/recommendations"
	"github.com/oullin/internal/shared/endpoint"
	"github.com/oullin/internal/shared/portal"
	"github.com/oullin/internal/signatures"
	"github.com/oullin/internal/talks"
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
	abstract := posts.NewPostsHandler(&repo)

	index := r.PipelineFor(abstract.Index)
	show := r.PipelineFor(abstract.Show)

	r.Mux.HandleFunc("POST /posts", index)
	r.Mux.HandleFunc("GET /posts/{slug}", show)
}

func (r *Router) Categories() {
	repo := repository.Categories{DB: r.Db}
	abstract := categories.NewCategoriesHandler(&repo)

	index := r.PipelineFor(abstract.Index)

	r.Mux.HandleFunc("GET /categories", index)
}

func (r *Router) Signature() {
	abstract := signatures.NewSignaturesHandler(r.Validator, r.Pipeline.ApiKeys)
	generate := r.PublicPipelineFor(abstract.Generate)

	r.Mux.HandleFunc("POST /generate-signature", generate)
}

func (r *Router) KeepAlive() {
	abstract := health.NewKeepAliveHandler(&r.Env.Ping)

	apiHandler := endpoint.NewApiHandler(
		r.Pipeline.Chain(abstract.Handle),
	)

	r.Mux.HandleFunc("GET /ping", apiHandler)
}

func (r *Router) KeepAliveDB() {
	abstract := health.NewKeepAliveDBHandler(&r.Env.Ping, r.Db)

	apiHandler := endpoint.NewApiHandler(
		r.Pipeline.Chain(abstract.Handle),
	)

	r.Mux.HandleFunc("GET /ping-db", apiHandler)
}

func (r *Router) Metrics() {
	metricsHandler := metrics.NewMetricsHandler()

	// Metrics endpoint blocked from public access by Caddy (see @protected matcher in Caddyfile)
	// Only accessible internally via direct container access (api:8080/metrics)
	// Prometheus scrapes via internal DNS without going through Caddy's public listener
	r.Mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, req *http.Request) {
		_ = metricsHandler.Handle(w, req)
	})
}

func (r *Router) Profile() {
	maker := profile.NewProfileHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetProfile(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Experience() {
	maker := experience.NewExperienceHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetExperience(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Projects() {
	maker := projects.NewProjectsHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetProjects(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Links() {
	maker := links.NewLinksHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetLinks(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Talks() {
	maker := talks.NewTalksHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetTalks(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Education() {
	maker := education.NewEducationHandlerWithCache

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetEducation(),
		func(file string, cacheEnabled bool) StaticRouteResource {
			return maker(file, cacheEnabled)
		},
	)
}

func (r *Router) Recommendations() {
	maker := recommendations.NewRecommendationsHandlerWithCache

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
