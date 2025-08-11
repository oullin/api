package kernel

import (
	baseHttp "net/http"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/handler"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/middleware"
)

type StaticRouteResource interface {
	Handle(baseHttp.ResponseWriter, *baseHttp.Request) *http.ApiError
}

func addStaticRoute[H StaticRouteResource](r *Router, path, file string, maker func(string) H) {
	abstract := maker(file)
	resolver := r.PipelineFor(abstract.Handle)
	r.Mux.HandleFunc("GET "+path, resolver)
}

type Router struct {
	Env      *env.Environment
	Mux      *baseHttp.ServeMux
	Pipeline middleware.Pipeline
	Db       *database.Connection
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

func (r *Router) Profile() {
	addStaticRoute(r, "/profile", "./storage/fixture/profile.json", handler.MakeProfileHandler)
}

func (r *Router) Experience() {
	addStaticRoute(r, "/experience", "./storage/fixture/experience.json", handler.MakeExperienceHandler)
}

func (r *Router) Projects() {
	addStaticRoute(r, "/projects", "./storage/fixture/projects.json", handler.MakeProjectsHandler)
}

func (r *Router) Social() {
	addStaticRoute(r, "/social", "./storage/fixture/social.json", handler.MakeSocialHandler)
}

func (r *Router) Talks() {
	addStaticRoute(r, "/talks", "./storage/fixture/talks.json", handler.MakeTalksHandler)
}

func (r *Router) Education() {
	addStaticRoute(r, "/education", "./storage/fixture/education.json", handler.MakeEducationHandler)
}

func (r *Router) Recommendations() {
	addStaticRoute(r, "/recommendations", "./storage/fixture/recommendations.json", handler.MakeRecommendationsHandler)
}
