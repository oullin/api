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
	return endpoint.MakeApiHandler(
		r.Pipeline.Chain(
			apiHandler,
			r.Pipeline.PublicMiddleware.Handle,
		),
	)
}

func (r *Router) PipelineFor(apiHandler endpoint.ApiHandler) http.HandlerFunc {
	tokenMiddleware := middleware.MakeTokenMiddleware(
		r.Pipeline.TokenHandler,
		r.Pipeline.ApiKeys,
	)

	return endpoint.MakeApiHandler(
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
	abstract := handler.MakeSignaturesHandler(r.Validator, r.Pipeline.ApiKeys)
	generate := r.PublicPipelineFor(abstract.Generate)

	r.Mux.HandleFunc("POST /generate-signature", generate)
}

func (r *Router) KeepAlive() {
	abstract := handler.MakeKeepAliveHandler(&r.Env.Ping)

	apiHandler := endpoint.MakeApiHandler(
		r.Pipeline.Chain(abstract.Handle),
	)

	r.Mux.HandleFunc("GET /ping", apiHandler)
}

func (r *Router) KeepAliveDB() {
	abstract := handler.MakeKeepAliveDBHandler(&r.Env.Ping, r.Db)

	apiHandler := endpoint.MakeApiHandler(
		r.Pipeline.Chain(abstract.Handle),
	)

	r.Mux.HandleFunc("GET /ping-db", apiHandler)
}

func (r *Router) Profile() {
	maker := handler.MakeProfileHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetProfile(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Experience() {
	maker := handler.MakeExperienceHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetExperience(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Projects() {
	maker := handler.MakeProjectsHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetProjects(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Social() {
	maker := handler.MakeSocialHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetSocial(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Talks() {
	maker := handler.MakeTalksHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetTalks(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Education() {
	maker := handler.MakeEducationHandler

	r.composeFixtures(
		r.WebsiteRoutes.Fixture.GetEducation(),
		func(file string) StaticRouteResource {
			return maker(file)
		},
	)
}

func (r *Router) Recommendations() {
	maker := handler.MakeRecommendationsHandler

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
