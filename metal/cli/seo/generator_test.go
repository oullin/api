package seo

import (
	"encoding/json"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"

	"github.com/oullin/database"
	"github.com/oullin/pkg/portal"
)

func newTestValidator(t *testing.T) *portal.Validator {
	t.Helper()

	return portal.MakeValidatorFrom(validator.New(validator.WithRequiredStructEnabled()))
}

func TestGeneratorBuildAndExport(t *testing.T) {
	page := Page{
		SiteName:      "SEO Test Suite",
		SiteURL:       "https://seo.example.test",
		Lang:          "en_GB",
		AboutPhotoUrl: "https://seo.example.test/photo.png",
		LogoURL:       "https://seo.example.test/logo.png",
		SameAsURL:     []string{"https://github.com/oullin"},
		Categories:    []string{"golang"},
		StubPath:      StubPath,
		OutputDir:     t.TempDir(),
	}

	tmpl, err := page.Load()
	if err != nil {
		t.Fatalf("load template: %v", err)
	}

	page.Template = tmpl

	gen := &Generator{
		Page:      page,
		Validator: newTestValidator(t),
	}

	body := []template.HTML{"<h1>Profile</h1><p>hello</p>"}
	data, err := gen.buildForPage(WebHomeName, WebHomeUrl, body)
	if err != nil {
		t.Fatalf("build err: %v", err)
	}

	if string(data.JsonLD) == "" {
		t.Fatalf("expected jsonld data")
	}

	if len(data.Body) != 1 || data.Body[0] != body[0] {
		t.Fatalf("unexpected body slice: %#v", data.Body)
	}

	var manifest map[string]any
	if err := json.Unmarshal([]byte(data.Manifest), &manifest); err != nil {
		t.Fatalf("manifest parse: %v", err)
	}

	if manifest["short_name"].(string) != "SEO Test Suite" {
		t.Fatalf("unexpected manifest short name: %v", manifest["short_name"])
	}

	if err := gen.Export("index", data); err != nil {
		t.Fatalf("export err: %v", err)
	}

	output := filepath.Join(page.OutputDir, "index.seo.html")
	raw, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(raw)
	if !strings.Contains(content, "<h1>Profile</h1><p>hello</p>") {
		t.Fatalf("expected body content rendered, got %q", content)
	}

	if !strings.Contains(content, "<link rel=\"manifest\"") {
		t.Fatalf("expected manifest link in template")
	}
}

func TestGeneratorBuildRejectsInvalidTemplateData(t *testing.T) {
	gen := &Generator{
		Page: Page{
			SiteName:      "SEO Test Suite",
			SiteURL:       "invalid-url",
			Lang:          "en_GB",
			AboutPhotoUrl: "https://seo.example.test/photo.png",
			LogoURL:       "https://seo.example.test/logo.png",
			Categories:    []string{"golang"},
		},
		Validator: newTestValidator(t),
	}

	if _, err := gen.buildForPage(WebHomeName, WebHomeUrl, []template.HTML{"<p>hello</p>"}); err == nil || !strings.Contains(err.Error(), "invalid template data") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestGeneratorGenerateAllPages(t *testing.T) {
	withRepoRoot(t)

	conn, env := newPostgresConnection(t, &database.Category{})

	seedCategory(t, conn, "golang", "GoLang")
	seedCategory(t, conn, "cli", "CLI Tools")

	gen, err := NewGenerator(conn, env, newTestValidator(t))
	if err != nil {
		t.Fatalf("new generator err: %v", err)
	}

	if len(gen.Page.Categories) == 0 {
		t.Fatalf("expected categories from database")
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("generate err: %v", err)
	}

	output := filepath.Join(env.Seo.SpaDir, "index.seo.html")
	raw, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	content := string(raw)
	if !strings.Contains(content, "<h1>Talks</h1>") {
		t.Fatalf("expected talks section in generated html")
	}

	if !strings.Contains(content, "cli tools") {
		t.Fatalf("expected categories to be rendered: %q", content)
	}

	aboutRaw, err := os.ReadFile(filepath.Join(env.Seo.SpaDir, "about.seo.html"))
	if err != nil {
		t.Fatalf("read about output: %v", err)
	}

	aboutContent := strings.ToLower(string(aboutRaw))
	if !strings.Contains(aboutContent, "<h1>social</h1>") {
		t.Fatalf("expected social section in about page: %q", aboutContent)
	}

	if !strings.Contains(aboutContent, "<h1>recommendations</h1>") {
		t.Fatalf("expected recommendations section in about page: %q", aboutContent)
	}

	projectsRaw, err := os.ReadFile(filepath.Join(env.Seo.SpaDir, "projects.seo.html"))
	if err != nil {
		t.Fatalf("read projects output: %v", err)
	}

	projectsContent := strings.ToLower(string(projectsRaw))
	if !strings.Contains(projectsContent, "<h1>projects</h1>") {
		t.Fatalf("expected projects section in projects page: %q", projectsContent)
	}

	resumeRaw, err := os.ReadFile(filepath.Join(env.Seo.SpaDir, "resume.seo.html"))
	if err != nil {
		t.Fatalf("read resume output: %v", err)
	}

	resumeContent := strings.ToLower(string(resumeRaw))
	if !strings.Contains(resumeContent, "<h1>experience</h1>") {
		t.Fatalf("expected experience section in resume page: %q", resumeContent)
	}

	if !strings.Contains(resumeContent, "<h1>education</h1>") {
		t.Fatalf("expected education section in resume page: %q", resumeContent)
	}
}
