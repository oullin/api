package handler

import (
	"net/http"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestSocialHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/social", []map[string]string{{"uuid": "1"}}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeSocialHandler(p)
		return h
	})
}
