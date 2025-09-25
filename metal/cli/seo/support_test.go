package seo

import (
	"encoding/base64"
	"html/template"
	"testing"
)

func TestManifestDataURL(t *testing.T) {
	manifest := template.JS(`{"name":"app"}`)
	got := ManifestDataURL(manifest)

	encoded := base64.StdEncoding.EncodeToString([]byte(manifest))
	want := template.URL("data:application/manifest+json;base64," + encoded)

	if got != want {
		t.Fatalf("unexpected manifest data url: got %q want %q", got, want)
	}
}
