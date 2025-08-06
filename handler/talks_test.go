package handler

import (
	"net/http"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestTalksHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/talks", []map[string]string{{"uuid": "1"}}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeTalksHandler(p)
		return h
	})
}
