package seo

import (
	"encoding/base64"
	"html/template"
)

func ManifestDataURL(manifest template.JS) template.URL {
	b64 := base64.StdEncoding.EncodeToString([]byte(manifest))
	u := "data:application/manifest+json;base64," + b64

	return template.URL(u)
}
