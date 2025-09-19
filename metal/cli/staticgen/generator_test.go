package staticgen

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/oullin/metal/kernel"
)

func testAssetConfig() AssetConfig {
	return AssetConfig{
		BuildRev:      "build-123",
		CanonicalBase: "https://example.com",
		DefaultLang:   "en",
		SiteName:      "Test Site",
	}
}

func TestGenerator_GenerateCreatesFiles(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", ".."))

	outputDir := t.TempDir()
	assets := testAssetConfig()
	generator, err := NewGenerator(outputDir, assets)
	if err != nil {
		t.Fatalf("unexpected error creating generator: %v", err)
	}

	routes := kernel.StaticRouteDefinitions()
	for i := range routes {
		if routes[i].File != "" && !filepath.IsAbs(routes[i].File) {
			routes[i].File = filepath.Join(repoRoot, routes[i].File)
		}
	}
	if len(routes) > 0 {
		routes[0].Page.Robots = "noindex,nofollow"
		routes[0].Page.ThemeColor = "#111827"
		routes[0].Page.JsonLD = `{"@context":"https://schema.org","@type":"WebPage"}`
		routes[0].Page.OG = kernel.OGSpec{
			Type:       "article",
			Image:      "https://cdn.example.com/profile.jpg",
			ImageAlt:   "Profile preview",
			ImageWidth: "1200",
		}
		routes[0].Page.Twitter = kernel.TwitterSpec{
			Card:     "summary",
			Image:    "https://cdn.example.com/profile.jpg",
			ImageAlt: "Profile preview",
		}
		routes[0].Page.Hreflangs = []kernel.Hreflang{
			{Lang: "en", Href: "https://example.com/profile"},
			{Lang: "es", Href: "https://example.com/es/profile"},
			{Lang: "", Href: ""},
		}
		routes[0].Page.Favicons = []kernel.Favicon{
			{Rel: "icon", Href: "/favicon.ico", Type: "image/x-icon"},
			{Rel: "icon", Href: "/icon-192.png", Type: "image/png", Sizes: "192x192"},
			{Rel: "", Href: "/ignored.png"},
		}
		routes[0].Page.Manifest = "/manifest.json"
		routes[0].Page.AppleTouchIcon = "/apple-touch-icon.png"
	}
	files, err := generator.Generate(routes)
	if err != nil {
		t.Fatalf("expected no error generating static routes, got %v", err)
	}

	if len(files) != len(routes) {
		t.Fatalf("expected %d files, got %d", len(routes), len(files))
	}

	for i, file := range files {
		route := routes[i]

		if _, err := os.Stat(file); err != nil {
			t.Fatalf("expected file %s to exist: %v", file, err)
		}

		expectedPath := generator.filePathFor(route.Path)
		if filepath.Clean(file) != filepath.Clean(expectedPath) {
			t.Fatalf("unexpected file path for %s: expected %s got %s", route.Path, expectedPath, file)
		}

		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed reading generated file %s: %v", file, err)
		}

		if len(contents) == 0 {
			t.Fatalf("generated file %s was empty", file)
		}

		htmlBody := string(contents)
		data := generator.templateData(route)
		lang := generator.langFor(route)

		if !strings.Contains(htmlBody, "<!doctype html>") {
			t.Fatalf("generated HTML for %s missing doctype", route.Path)
		}

		if !strings.Contains(htmlBody, fmt.Sprintf("<html lang=\"%s\">", lang)) {
			t.Fatalf("expected lang attribute %q in %s", lang, file)
		}

		if !strings.Contains(htmlBody, fmt.Sprintf("<title>%s</title>", html.EscapeString(data.Title))) {
			t.Fatalf("expected title %q in %s", data.Title, file)
		}

		escapedDescription := html.EscapeString(data.Description)
		if !strings.Contains(htmlBody, fmt.Sprintf("<meta name=\"description\" content=\"%s\">", escapedDescription)) {
			t.Fatalf("expected description meta for %s", file)
		}

		expectedRobots := data.Robots
		if expectedRobots == "" {
			expectedRobots = "index,follow"
		}
		if !strings.Contains(htmlBody, fmt.Sprintf("<meta name=\"robots\" content=\"%s\">", html.EscapeString(expectedRobots))) {
			t.Fatalf("expected robots meta for %s", file)
		}

		if !strings.Contains(htmlBody, "<meta name=\"referrer\" content=\"strict-origin-when-cross-origin\">") {
			t.Fatalf("expected referrer policy meta for %s", file)
		}

		if !strings.Contains(htmlBody, "<meta name=\"color-scheme\" content=\"light dark\">") {
			t.Fatalf("expected color-scheme meta for %s", file)
		}

		if data.ThemeColor != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<meta name=\"theme-color\" content=\"%s\">", html.EscapeString(data.ThemeColor))) {
				t.Fatalf("expected theme color meta for %s", file)
			}
		}

		if !strings.Contains(htmlBody, fmt.Sprintf("<meta name=\"x-build-rev\" content=\"%s\">", html.EscapeString(data.BuildRev))) {
			t.Fatalf("expected build rev meta for %s", file)
		}

		if data.Canonical != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<link rel=\"canonical\" href=\"%s\">", html.EscapeString(data.Canonical))) {
				t.Fatalf("expected canonical link for %s", file)
			}
		}

		expectedOGType := data.OG.Type
		if expectedOGType == "" {
			expectedOGType = "website"
		}
		if !strings.Contains(htmlBody, fmt.Sprintf("<meta property=\"og:type\" content=\"%s\">", html.EscapeString(expectedOGType))) {
			t.Fatalf("expected og:type meta for %s", file)
		}

		if data.OG.Image != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<meta property=\"og:image\" content=\"%s\">", html.EscapeString(data.OG.Image))) {
				t.Fatalf("expected og:image meta for %s", file)
			}
		}

		expectedOGSiteName := data.OG.SiteName
		if expectedOGSiteName == "" {
			expectedOGSiteName = assets.SiteName
		}
		if expectedOGSiteName != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<meta property=\"og:site_name\" content=\"%s\">", html.EscapeString(expectedOGSiteName))) {
				t.Fatalf("expected og:site_name meta for %s", file)
			}
		}

		if data.OG.Locale != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<meta property=\"og:locale\" content=\"%s\">", html.EscapeString(data.OG.Locale))) {
				t.Fatalf("expected og:locale meta for %s", file)
			}
		}

		expectedTwitterCard := data.Twitter.Card
		if expectedTwitterCard == "" {
			expectedTwitterCard = "summary_large_image"
		}
		if !strings.Contains(htmlBody, fmt.Sprintf("<meta name=\"twitter:card\" content=\"%s\">", html.EscapeString(expectedTwitterCard))) {
			t.Fatalf("expected twitter:card meta for %s", file)
		}

		if data.Twitter.Image != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<meta name=\"twitter:image\" content=\"%s\">", html.EscapeString(data.Twitter.Image))) {
				t.Fatalf("expected twitter:image meta for %s", file)
			}
		}

		jsonLD := strings.TrimSpace(string(data.JsonLD))
		if jsonLD != "" {
			scriptTag := fmt.Sprintf("<script type=\"application/ld+json\">%s</script>", jsonLD)
			if !strings.Contains(htmlBody, scriptTag) {
				t.Fatalf("expected json-ld script for %s", file)
			}
		}

		for _, hreflang := range data.Hreflangs {
			expected := fmt.Sprintf("<link rel=\"alternate\" hreflang=\"%s\" href=\"%s\">", html.EscapeString(hreflang.Lang), html.EscapeString(hreflang.Href))
			if !strings.Contains(htmlBody, expected) {
				t.Fatalf("expected hreflang link %s for %s", expected, file)
			}
		}

		for _, favicon := range data.Favicons {
			builder := strings.Builder{}
			builder.WriteString("<link rel=\"")
			builder.WriteString(html.EscapeString(favicon.Rel))
			builder.WriteString("\" href=\"")
			builder.WriteString(html.EscapeString(favicon.Href))
			builder.WriteString("\"")
			if favicon.Type != "" {
				builder.WriteString(" type=\"")
				builder.WriteString(html.EscapeString(favicon.Type))
				builder.WriteString("\"")
			}
			if favicon.Sizes != "" {
				builder.WriteString(" sizes=\"")
				builder.WriteString(html.EscapeString(favicon.Sizes))
				builder.WriteString("\"")
			}
			builder.WriteString(">")

			if !strings.Contains(htmlBody, builder.String()) {
				t.Fatalf("expected favicon link %s for %s", builder.String(), file)
			}
		}

		if data.Manifest != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<link rel=\"manifest\" href=\"%s\">", html.EscapeString(data.Manifest))) {
				t.Fatalf("expected manifest link for %s", file)
			}
		}

		if data.AppleTouchIcon != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<link rel=\"apple-touch-icon\" href=\"%s\">", html.EscapeString(data.AppleTouchIcon))) {
				t.Fatalf("expected apple-touch-icon link for %s", file)
			}
		}

		if !strings.HasPrefix(filepath.Clean(file), filepath.Clean(outputDir)) {
			t.Fatalf("generated file %s was outside the output directory", file)
		}
	}
}

func TestGenerator_GenerateRequiresOutputDir(t *testing.T) {
	t.Parallel()

	generator, err := NewGenerator(" ", testAssetConfig())
	if err != nil {
		t.Fatalf("unexpected error creating generator: %v", err)
	}

	if _, err := generator.Generate(kernel.StaticRouteDefinitions()); err == nil {
		t.Fatal("expected an error when no output directory is provided")
	}
}

func TestGenerator_GenerateFailsWithoutHandler(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	generator, err := NewGenerator(outputDir, testAssetConfig())
	if err != nil {
		t.Fatalf("unexpected error creating generator: %v", err)
	}

	routes := []kernel.StaticRouteDefinition{
		{
			Path:  "/invalid",
			Maker: func(string) kernel.StaticRouteResource { return nil },
		},
	}

	if _, err := generator.Generate(routes); err == nil {
		t.Fatal("expected an error when a route does not provide a handler")
	}
}
