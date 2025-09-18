package staticgen

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/metal/kernel"
)

func testAssetConfig() AssetConfig {
	return AssetConfig{
		BuildRev:      "build-123",
		AppCSS:        "/static/app.css",
		CanonicalBase: "https://example.com",
		DefaultLang:   "en",
		SiteName:      "Test Site",
	}
}

func TestGenerator_GenerateCreatesFiles(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed getting working directory: %v", err)
	}

	repoRoot := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	if err := os.Chdir(repoRoot); err != nil {
		t.Fatalf("failed changing directory to repository root: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("failed restoring working directory: %v", err)
		}
	})

	outputDir := t.TempDir()
	assets := testAssetConfig()
	generator, err := NewGenerator(outputDir, assets)
	if err != nil {
		t.Fatalf("unexpected error creating generator: %v", err)
	}

	routes := kernel.StaticRouteDefinitions()
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

		if !strings.Contains(htmlBody, fmt.Sprintf("<meta name=\"x-build-rev\" content=\"%s\">", html.EscapeString(data.BuildRev))) {
			t.Fatalf("expected build rev meta for %s", file)
		}

		if data.Canonical != "" {
			if !strings.Contains(htmlBody, fmt.Sprintf("<link rel=\"canonical\" href=\"%s\">", html.EscapeString(data.Canonical))) {
				t.Fatalf("expected canonical link for %s", file)
			}
		}

		if data.AppCSS != "" {
			expectedStylesheet := fmt.Sprintf("<link rel=%q href=%q>", "stylesheet", html.EscapeString(data.AppCSS))
			if !strings.Contains(htmlBody, expectedStylesheet) {
				t.Fatalf("expected stylesheet link for %s", file)
			}
		}

		if !strings.Contains(htmlBody, "hello world") {
			t.Fatalf("expected hello world in body for %s", file)
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
