package seo_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/oullin/internal/seo"
)

func TestJsonIDRenderProducesGraph(t *testing.T) {
	id := (&seo.JsonID{
		SiteURL:         "https://example.test",
		OrgName:         "Example",
		LogoURL:         "https://example.test/logo.png",
		Lang:            "en_GB",
		FoundedYear:     "2020",
		SameAs:          []string{"https://github.com/example"},
		SiteDescription: "Example description",
		Now: func() time.Time {
			return time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
		},
	}).
		WithPage("About", "AboutPage", "https://example.test/about", "About Example").
		WithFounder(seo.JsonPerson{
			Name:        "Jane Example",
			JobTitle:    "Founder of Example",
			URL:         "https://example.test/about",
			Description: "Founder bio",
		})

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

	if len(graph) != 4 {
		t.Fatalf("expected graph entries, got %d", len(graph))
	}

	org := graph[0].(map[string]any)
	if org["@id"].(string) != "https://example.test#org" {
		t.Fatalf("unexpected org id %q", org["@id"])
	}

	if org["founder"].(map[string]any)["@id"].(string) != "https://example.test#founder" {
		t.Fatalf("unexpected org founder ref %#v", org["founder"])
	}

	page := graph[2].(map[string]any)
	if page["@type"].(string) != "AboutPage" {
		t.Fatalf("expected AboutPage entry, got %q", page["@type"])
	}

	if _, ok := page["@context"]; ok {
		t.Fatalf("page should inherit root context: %#v", page)
	}

	if page["founder"].(map[string]any)["@id"].(string) != "https://example.test#founder" {
		t.Fatalf("unexpected page founder ref %#v", page["founder"])
	}

	founder := graph[3].(map[string]any)
	if founder["@type"].(string) != "Person" {
		t.Fatalf("expected Person entry, got %q", founder["@type"])
	}

	if _, ok := founder["@context"]; ok {
		t.Fatalf("founder should inherit root context: %#v", founder)
	}

	if founder["@id"].(string) != "https://example.test#founder" {
		t.Fatalf("unexpected founder id %q", founder["@id"])
	}
}

func TestJsonIDRenderWithoutFounderProducesBrandOnlyGraph(t *testing.T) {
	id := (&seo.JsonID{
		SiteURL:         "https://oullin.io",
		OrgName:         "Oullin",
		LogoURL:         "https://oullin.io/logo.png",
		Lang:            "en_GB",
		FoundedYear:     "2020",
		SameAs:          []string{"https://github.com/oullin"},
		SiteDescription: "Oullin description",
		Now: func() time.Time {
			return time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
		},
	}).WithPage("About", "AboutPage", "https://oullin.io/about", "About Oullin")

	rendered := id.Render()

	var got map[string]any
	if err := json.Unmarshal([]byte(rendered), &got); err != nil {
		t.Fatalf("jsonld parse err: %v", err)
	}

	graph := got["@graph"].([]any)
	if len(graph) != 3 {
		t.Fatalf("expected 3 graph entries without founder, got %d", len(graph))
	}

	org := graph[0].(map[string]any)
	if org["name"].(string) != "Oullin" {
		t.Fatalf("expected organization name to match brand, got %#v", org)
	}
	if _, ok := org["founder"]; ok {
		t.Fatalf("did not expect founder on org: %#v", org)
	}

	page := graph[2].(map[string]any)
	if page["name"].(string) != "About" {
		t.Fatalf("unexpected page name %#v", page)
	}
	if _, ok := page["founder"]; ok {
		t.Fatalf("did not expect founder on page: %#v", page)
	}
}
