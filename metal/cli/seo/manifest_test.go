package seo

import (
	"encoding/json"
	"html/template"
	"testing"
	"time"
)

func TestManifestRenderUsesFavicons(t *testing.T) {
	tmpl := Page{
		SiteName:   "Example Site",
		SiteURL:    "https://example.test",
		Lang:       "en_GB",
		LogoURL:    "https://example.test/logo.png",
		SameAsURL:  []string{"https://example.test"},
		StubPath:   StubPath,
		OutputDir:  t.TempDir(),
		Categories: []string{"go"},
	}

	data := TemplateData{
		Lang:        "en_GB",
		Title:       "Example Site",
		Description: "Example Site description",
		Canonical:   "https://example.test",
		Robots:      "index,follow",
		ThemeColor:  "#fff",
		JsonLD:      template.JS("{}"),
		OGTagOg: TagOgData{
			Type:        "website",
			Image:       "https://example.test/logo.png",
			ImageAlt:    "Example Site",
			ImageWidth:  "600",
			ImageHeight: "400",
			SiteName:    "Example Site",
			Locale:      "en_GB",
		},
		Twitter: TwitterData{
			Card:     "summary_large_image",
			Image:    "https://example.test/logo.png",
			ImageAlt: "Example Site",
		},
		HrefLang: []HrefLangData{{Lang: "en_GB", Href: "https://example.test"}},
		Favicons: []FaviconData{{
			Rel:   "icon",
			Href:  "https://example.test/favicon.ico",
			Type:  "image/x-icon",
			Sizes: "48x48",
		}},
		Manifest:       template.JS("{}"),
		AppleTouchIcon: "https://example.test/apple.png",
		Categories:     []string{"go"},
		BgColor:        "#000",
		Body:           []template.HTML{"<p>body</p>"},
	}

	manifest := NewManifest(tmpl, data)
	manifest.Now = func() time.Time { return time.Unix(0, 0).UTC() }

	rendered := manifest.Render()

	var got map[string]any
	if err := json.Unmarshal([]byte(rendered), &got); err != nil {
		t.Fatalf("manifest JSON parse err: %v", err)
	}

	icons, ok := got["icons"].([]any)
	if !ok || len(icons) != 1 {
		t.Fatalf("expected 1 icon got %#v", got["icons"])
	}

	icon := icons[0].(map[string]any)
	if icon["src"].(string) != "https://example.test/favicon.ico" {
		t.Fatalf("unexpected icon src %q", icon["src"])
	}

	if got["shortcuts"] == nil {
		t.Fatalf("expected shortcuts in manifest")
	}
}

func TestManifestRenderFallsBackToLogo(t *testing.T) {
	tmpl := Page{
		SiteName:   "Fallback",
		SiteURL:    "https://fallback.test",
		Lang:       "en_GB",
		LogoURL:    "https://fallback.test/logo.png",
		SameAsURL:  []string{"https://fallback.test"},
		StubPath:   StubPath,
		OutputDir:  t.TempDir(),
		Categories: []string{"go"},
	}

	data := TemplateData{
		Lang:        "en_GB",
		Title:       "Fallback",
		Description: "Fallback description",
		Canonical:   "https://fallback.test",
		Robots:      "index,follow",
		ThemeColor:  "#fff",
		JsonLD:      template.JS("{}"),
		OGTagOg: TagOgData{
			Type:        "website",
			Image:       "https://fallback.test/logo.png",
			ImageAlt:    "Fallback",
			ImageWidth:  "600",
			ImageHeight: "400",
			SiteName:    "Fallback",
			Locale:      "en_GB",
		},
		Twitter: TwitterData{
			Card:     "summary_large_image",
			Image:    "https://fallback.test/logo.png",
			ImageAlt: "Fallback",
		},
		HrefLang:       []HrefLangData{{Lang: "en_GB", Href: "https://fallback.test"}},
		Favicons:       nil,
		Manifest:       template.JS("{}"),
		AppleTouchIcon: "https://fallback.test/apple.png",
		Categories:     []string{"go"},
		BgColor:        "#000",
		Body:           []template.HTML{"<p>body</p>"},
	}

	manifest := NewManifest(tmpl, data)
	rendered := manifest.Render()

	var got map[string]any
	if err := json.Unmarshal([]byte(rendered), &got); err != nil {
		t.Fatalf("manifest JSON parse err: %v", err)
	}

	icons := got["icons"].([]any)
	icon := icons[0].(map[string]any)
	if icon["src"].(string) != "https://fallback.test/logo.png" {
		t.Fatalf("expected fallback logo src, got %q", icon["src"])
	}
}
