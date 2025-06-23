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

func MakeRouter(mux *baseHttp.ServeMux) *Router {
	return &Router{
		Mux: mux,
	}
}

func (r *Router) Profile(fixture string) {
	tokenMiddleware := middleware.MakeTokenMiddleware(
		r.Env.App.Credentials,
	)

	profileHandler := handler.ProfileHandler{
		Fixture: fixture,
	}

	getHandler := http.MakeApiHandler(
		r.Pipeline.Chain(
			profileHandler.Handle,
			middleware.UsernameCheck,
			tokenMiddleware.Handle,
		),
	)

	r.Mux.HandleFunc("GET /profile", getHandler)
}
