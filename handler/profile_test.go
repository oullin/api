package handler

import (
	"net/http"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestProfileHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/profile", map[string]string{"nickname": "nick"}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeProfileHandler(p)
		return h
	})
}
