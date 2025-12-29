package seo_test

import (
	"encoding/base64"
	"html/template"
	"testing"

	"github.com/oullin/metal/cli/seo"
)

func TestManifestDataURL(t *testing.T) {
	manifest := template.JS(`{"name":"app"}`)
	got := seo.ManifestDataURL(manifest)

	encoded := base64.StdEncoding.EncodeToString([]byte(manifest))
	want := template.URL("data:application/manifest+json;base64," + encoded)

	if got != want {
		t.Fatalf("unexpected manifest data url: got %q want %q", got, want)
	}
}
