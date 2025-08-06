package handler

import (
	"net/http"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestExperienceHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/experience", []map[string]string{{"uuid": "1"}}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeExperienceHandler(p)
		return h
	})
}
