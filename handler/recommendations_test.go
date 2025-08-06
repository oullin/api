package handler

import (
	"net/http"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestRecommendationsHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/recommendations", []map[string]string{{"uuid": "1"}}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeRecommendationsHandler(p)
		return h
	})
}
