package boost

import (
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
}

func (r *Router) PipelineFor(apiHandler http.ApiHandler) baseHttp.HandlerFunc {
	//tokenMiddleware := middleware.MakeTokenMiddleware(
	//	r.Env.App.Credentials,
	//)

	return http.MakeApiHandler(
		r.Pipeline.Chain(
			apiHandler,
			middleware.UsernameCheck,
			//tokenMiddleware.Handle,
		),
	)
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
