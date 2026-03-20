package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oullin/internal/app/middleware"

	"github.com/oullin/internal/shared/endpoint"
)

func TestPipelineChainOrder(t *testing.T) {
	p := middleware.Pipeline{}

	order := []string{}

	m1 := func(next endpoint.ApiHandler) endpoint.ApiHandler {

		return func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
			order = append(order, "m1")

			return next(w, r)
		}
	}

	m2 := func(next endpoint.ApiHandler) endpoint.ApiHandler {

		return func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
			order = append(order, "m2")

			return next(w, r)
		}
	}

	final := func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
		order = append(order, "final")

		return nil
	}

	chained := p.Chain(final, m1, m2)
	chained(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

	joined := strings.Join(order, ",")

	if joined != "m1,m2,final" {
		t.Fatalf("order wrong: %s", joined)
	}
}
