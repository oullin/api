package seo

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"net/http/httptest"

	"github.com/oullin/metal/router"
)

func Fetch[T any](response *T, handler func() router.StaticRouteResource) error {
	req := httptest.NewRequest("GET", "http://localhost:8080/proxy", nil)
	rr := httptest.NewRecorder()

	maker := handler()

	if err := maker.Handle(rr, req); err != nil {
		return err
	}

	if err := json.Unmarshal(rr.Body.Bytes(), response); err != nil {
		return err
	}

	return nil
}

func ManifestDataURL(manifest template.JS) template.URL {
	b64 := base64.StdEncoding.EncodeToString([]byte(manifest))
	u := "data:application/manifest+json;base64," + b64

	return template.URL(u)
}
