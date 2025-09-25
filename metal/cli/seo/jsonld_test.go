package seo

import (
	"encoding/json"
	"testing"
	"time"
)

func TestJsonIDRenderProducesGraph(t *testing.T) {
	id := &JsonID{
		SiteURL:     "https://example.test",
		OrgName:     "Example",
		LogoURL:     "https://example.test/logo.png",
		Lang:        "en_GB",
		FoundedYear: "2020",
		SameAs:      []string{"https://example.test", "https://github.com/example"},
		WebRepoURL:  "https://github.com/example/web",
		APIRepoURL:  "https://github.com/example/api",
		WebName:     "Example Web",
		APIName:     "Example API",
		Now: func() time.Time {
			return time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
		},
	}

	rendered := id.Render()

	var got map[string]any
	if err := json.Unmarshal([]byte(rendered), &got); err != nil {
		t.Fatalf("jsonld parse err: %v", err)
	}

	if got["@context"].(string) != "https://schema.org" {
		t.Fatalf("unexpected context %q", got["@context"])
	}

	graph, ok := got["@graph"].([]any)
	if !ok {
		t.Fatalf("missing graph: %#v", got["@graph"])
	}

	if len(graph) < 5 {
		t.Fatalf("expected graph entries, got %d", len(graph))
	}

	org := graph[0].(map[string]any)
	if org["@id"].(string) != "https://example.test#org" {
		t.Fatalf("unexpected org id %q", org["@id"])
	}

	api := graph[len(graph)-1].(map[string]any)
	if api["@type"].(string) != "WebAPI" {
		t.Fatalf("expected WebAPI entry, got %q", api["@type"])
	}
}
