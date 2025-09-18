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
		BuildRev:        "build-123",
		AuthBootstrapJS: "/static/auth.js",
		APIBase:         "https://api.example.com",
		SessionPath:     "/auth/session",
		LoginURL:        "https://example.com/login",
		AppJS:           "/static/app.js",
		AppCSS:          "/static/app.css",
		CanonicalBase:   "https://example.com",
		DefaultLang:     "en",
		SiteName:        "Test Site",
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

		expectedDir := generator.directoryFor(route.Path)
		expectedPath := filepath.Join(expectedDir, "index.html")
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
			if !strings.Contains(htmlBody, fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\">", html.EscapeString(data.AppCSS))) {
				t.Fatalf("expected stylesheet link for %s", file)
			}
		}

		if !strings.Contains(htmlBody, fmt.Sprintf("src=\"%s\"", html.EscapeString(data.AuthBootstrapJS))) {
			t.Fatalf("expected auth bootstrap script src for %s", file)
		}

		if !strings.Contains(htmlBody, fmt.Sprintf("data-api-base=\"%s\"", html.EscapeString(data.APIBase))) {
			t.Fatalf("expected api base data attribute for %s", file)
		}

		if !strings.Contains(htmlBody, fmt.Sprintf("data-session-path=\"%s\"", html.EscapeString(data.SessionPath))) {
			t.Fatalf("expected session path data attribute for %s", file)
		}

		expectedSession := fmt.Sprintf("data-session-path=\"%s\"%s", html.EscapeString(data.SessionPath), sessionComment)
		if !strings.Contains(htmlBody, expectedSession) {
			t.Fatalf("expected inline session comment for %s", file)
		}

		if !strings.Contains(htmlBody, fmt.Sprintf("data-login-url=\"%s\"", html.EscapeString(data.LoginURL))) {
			t.Fatalf("expected login url data attribute for %s", file)
		}

		expectedLogin := fmt.Sprintf("data-login-url=\"%s\"%s", html.EscapeString(data.LoginURL), loginComment)
		if !strings.Contains(htmlBody, expectedLogin) {
			t.Fatalf("expected inline login comment for %s", file)
		}

		if !strings.Contains(htmlBody, fmt.Sprintf("data-app-js=\"%s\"", html.EscapeString(data.AppJS))) {
			t.Fatalf("expected app js data attribute for %s", file)
		}

		expectedApp := fmt.Sprintf("data-app-js=\"%s\"%s", html.EscapeString(data.AppJS), appComment)
		if !strings.Contains(htmlBody, expectedApp) {
			t.Fatalf("expected inline app comment for %s", file)
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
