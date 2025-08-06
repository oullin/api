package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pkgHttp "github.com/oullin/pkg/http"
)

func TestPipelineChainOrder(t *testing.T) {
	p := Pipeline{}
	order := []string{}

	m1 := func(next pkgHttp.ApiHandler) pkgHttp.ApiHandler {

		return func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
			order = append(order, "m1")

			return next(w, r)
		}
	}
	m2 := func(next pkgHttp.ApiHandler) pkgHttp.ApiHandler {

		return func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
			order = append(order, "m2")

			return next(w, r)
		}
	}
	final := func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
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
