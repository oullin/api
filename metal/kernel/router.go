package kernel

import (
	baseHttp "net/http"
	"strings"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/handler"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/kernel/web"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/middleware"
	"github.com/oullin/pkg/portal"
)

const StaticRouteTalks = "talks"
const StaticRouteSocial = "social"
const StaticRouteProfile = "profile"
const StaticRouteProjects = "projects"
const StaticRouteEducation = "education"
const StaticRouteExperience = "experience"
const StaticRouteRecommendations = "recommendations"

func addStaticRoute[H web.StaticRouteResource](r *Router, path, file string, maker func(string) H) {
	abstract := maker(file)
	resolver := r.PipelineFor(abstract.Handle)

	r.WebsiteRoutes.AddPageFrom(path, file, func(file string) web.StaticRouteResource {
		return maker(file)
	})

	r.Mux.HandleFunc("GET "+path, resolver)
}

type Router struct {
	//@todo
	// --- make these fields required and use the validator to verify them.
	Env           *env.Environment
	Mux           *baseHttp.ServeMux
	Pipeline      middleware.Pipeline
	Db            *database.Connection
	validator     *portal.Validator
	WebsiteRoutes *web.Routes
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

func (r *Router) Profile() {
	path, file := r.StaticRouteFor(StaticRouteProfile)
	addStaticRoute(r, path, file, handler.MakeProfileHandler)
}

func (r *Router) Experience() {
	path, file := r.StaticRouteFor(StaticRouteExperience)
	addStaticRoute(r, path, file, handler.MakeExperienceHandler)
}

func (r *Router) Projects() {
	path, file := r.StaticRouteFor(StaticRouteProjects)
	addStaticRoute(r, path, file, handler.MakeProjectsHandler)
}

func (r *Router) Social() {
	path, file := r.StaticRouteFor(StaticRouteSocial)
	addStaticRoute(r, path, file, handler.MakeSocialHandler)
}

func (r *Router) Talks() {
	path, file := r.StaticRouteFor(StaticRouteTalks)
	addStaticRoute(r, path, file, handler.MakeTalksHandler)
}

func (r *Router) Education() {
	path, file := r.StaticRouteFor(StaticRouteEducation)
	addStaticRoute(r, path, file, handler.MakeEducationHandler)
}

func (r *Router) Recommendations() {
	maker := handler.MakeRecommendationsHandler
	path, file := r.StaticRouteFor(StaticRouteRecommendations)

	r.WebsiteRoutes.AddPageFrom(path, file, func(file string) web.StaticRouteResource {
		return maker(file)
	})

	addStaticRoute(r, path, file, maker)
}

func (r *Router) StaticRouteFor(slug string) (path string, file string) {
	filepath := "/" + strings.Trim(slug, "/")
	fixture := "./storage/fixture/" + slug + ".json"

	return filepath, fixture
}
