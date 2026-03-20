package middleware

import (
	"github.com/oullin/database/repository"
	env "github.com/oullin/internal/app/config"
	"github.com/oullin/internal/shared/auth"
	"github.com/oullin/internal/shared/endpoint"
)

type Pipeline struct {
	Env              *env.Environment
	ApiKeys          *repository.ApiKeys
	TokenHandler     *auth.TokenHandler //@todo Remove!
	PublicMiddleware PublicMiddleware
}

func (m Pipeline) Chain(h endpoint.ApiHandler, handlers ...endpoint.Middleware) endpoint.ApiHandler {
	for i := len(handlers) - 1; i >= 0; i-- {
		h = handlers[i](h)
	}

	return h
}
