package middleware

import (
    "github.com/oullin/env"
    "github.com/oullin/pkg/http"
)

type Pipeline struct {
    Env *env.Environment
}

// Chain applies a list of middleware handlers to a final ApiHandler.
// It builds the chain in reverse, so the first middleware
// in the list is the outermost one, executing first.
func (m Pipeline) Chain(h http.ApiHandler, handlers ...http.Middleware) http.ApiHandler {
    for i := len(handlers) - 1; i >= 0; i-- {
        h = handlers[i](h)
    }

    return h
}
