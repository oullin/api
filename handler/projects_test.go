package handler

import (
	"net/http"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestProjectsHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/projects", []map[string]string{{"uuid": "1"}}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeProjectsHandler(p)
		return h
	})
}
