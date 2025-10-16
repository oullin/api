package middleware

import (
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/endpoint"
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
