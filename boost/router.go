package boost

import (
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
	"github.com/oullin/handler"
	"github.com/oullin/pkg/http"
	"github.com/oullin/pkg/http/middleware"
	baseHttp "net/http"
)

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

	resolver := r.PipelineFor(
		abstract.Handle,
	)

	r.Mux.HandleFunc("/posts", resolver)
}

func (r *Router) Profile() {
	abstract := handler.MakeProfileHandler("./storage/fixture/profile.json")

	resolver := r.PipelineFor(
		abstract.Handle,
	)

	r.Mux.HandleFunc("GET /profile", resolver)
}

func (r *Router) Experience() {
	abstract := handler.MakeExperienceHandler("./storage/fixture/experience.json")

	resolver := r.PipelineFor(
		abstract.Handle,
	)

	r.Mux.HandleFunc("GET /experience", resolver)
}

func (r *Router) Projects() {
	abstract := handler.MakeProjectsHandler("./storage/fixture/projects.json")

	resolver := r.PipelineFor(
		abstract.Handle,
	)

	r.Mux.HandleFunc("GET /projects", resolver)
}

func (r *Router) Social() {
	abstract := handler.MakeSocialHandler("./storage/fixture/social.json")

	resolver := r.PipelineFor(
		abstract.Handle,
	)

	r.Mux.HandleFunc("GET /social", resolver)
}

func (r *Router) Talks() {
	abstract := handler.MakeTalksHandler("./storage/fixture/talks.json")

	resolver := r.PipelineFor(
		abstract.Handle,
	)

	r.Mux.HandleFunc("GET /talks", resolver)
}

func (r *Router) Education() {
	abstract := handler.MakeEducationHandler("./storage/fixture/education.json")

	resolver := r.PipelineFor(
		abstract.Handle,
	)

	r.Mux.HandleFunc("GET /education", resolver)
}

func (r *Router) Recommendations() {
	abstract := handler.MakeRecommendationsHandler("./storage/fixture/recommendations.json")

	resolver := r.PipelineFor(
		abstract.Handle,
	)

	r.Mux.HandleFunc("GET /recommendations", resolver)
}
