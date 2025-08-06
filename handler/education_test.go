package handler

import (
	"net/http"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestEducationHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/education", []map[string]string{{"uuid": "1"}}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeEducationHandler(p)
		return h
	})
}
