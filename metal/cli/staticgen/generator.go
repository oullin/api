package staticgen

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	baseHttp "net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/oullin/metal/kernel"
)

//go:embed templates/protected_index.oullin.html
var templatesFS embed.FS

const (
	protectedIndexTemplate = "templates/protected_index.oullin.html"
)

type AssetConfig struct {
	BuildRev      string
	AppCSS        string
	CanonicalBase string
	DefaultLang   string
	SiteName      string
}

type Generator struct {
	OutputDir string
	assets    AssetConfig
	tmpl      *htmltemplate.Template
}

type templateData struct {
	Lang        string
	Title       string
	Description string
	Canonical   string
	OG          ogData
	AppCSS      string
	BuildRev    string
}

type ogData struct {
	Image string
}

func NewGenerator(outputDir string, config AssetConfig) (Generator, error) {
	config = config.normalized()

	if err := config.validate(); err != nil {
		return Generator{}, err
	}

	tmpl, err := loadTemplate()
	if err != nil {
		return Generator{}, fmt.Errorf("parse template: %w", err)
	}

	return Generator{
		OutputDir: outputDir,
		assets:    config,
		tmpl:      tmpl,
	}, nil
}

func (c AssetConfig) normalized() AssetConfig {
	normalized := c

	normalized.BuildRev = strings.TrimSpace(normalized.BuildRev)
	normalized.AppCSS = strings.TrimSpace(normalized.AppCSS)
	normalized.CanonicalBase = strings.TrimRight(strings.TrimSpace(normalized.CanonicalBase), "/")
	normalized.DefaultLang = strings.TrimSpace(normalized.DefaultLang)
	normalized.SiteName = strings.TrimSpace(normalized.SiteName)

	if normalized.DefaultLang == "" {
		normalized.DefaultLang = "en"
	}

	return normalized
}

func (c AssetConfig) validate() error {
	missing := make([]string, 0, 1)

	if c.BuildRev == "" {
		missing = append(missing, "BuildRev")
	}

	if len(missing) > 0 {
		return fmt.Errorf("static generator configuration missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (g Generator) Generate(routes []kernel.StaticRouteDefinition) ([]string, error) {
	if strings.TrimSpace(g.OutputDir) == "" {
		return nil, fmt.Errorf("output directory must be provided")
	}

	if g.tmpl == nil {
		return nil, fmt.Errorf("template not configured")
	}

	if err := os.MkdirAll(g.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating output directory: %w", err)
	}

	generated := make([]string, 0, len(routes))

	for _, route := range routes {
		if strings.TrimSpace(route.Path) == "" {
			return nil, fmt.Errorf("static route path cannot be empty")
		}

		if route.Maker == nil {
			return nil, fmt.Errorf("static route %s is missing a handler", route.Path)
		}

		resource := route.Maker(route.File)
		if resource == nil {
			return nil, fmt.Errorf("static route %s returned no handler", route.Path)
		}

		request := httptest.NewRequest(baseHttp.MethodGet, route.Path, nil)
		recorder := httptest.NewRecorder()

		if apiErr := resource.Handle(recorder, request); apiErr != nil {
			return nil, fmt.Errorf("static route %s failed: %s", route.Path, apiErr.Message)
		}

		if recorder.Code != baseHttp.StatusOK {
			return nil, fmt.Errorf("static route %s returned status %d", route.Path, recorder.Code)
		}

		if recorder.Body.Len() == 0 {
			return nil, fmt.Errorf("static route %s returned an empty body", route.Path)
		}

		data := g.templateData(route)

		var buffer bytes.Buffer
		if err := g.tmpl.Execute(&buffer, data); err != nil {
			return nil, fmt.Errorf("rendering template for %s: %w", route.Path, err)
		}

		filePath := g.filePathFor(route.Path)

		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return nil, fmt.Errorf("creating directory for %s: %w", route.Path, err)
		}

		if err := os.WriteFile(filePath, buffer.Bytes(), 0o644); err != nil {
			return nil, fmt.Errorf("writing %s: %w", filePath, err)
		}

		generated = append(generated, filePath)
	}

	return generated, nil
}

func (g Generator) templateData(route kernel.StaticRouteDefinition) templateData {
	return templateData{
		Lang:        g.langFor(route),
		Title:       g.titleFor(route),
		Description: g.descriptionFor(route),
		Canonical:   g.canonicalFor(route),
		OG:          ogData{Image: strings.TrimSpace(route.Page.OGImage)},
		AppCSS:      g.assets.AppCSS,
		BuildRev:    g.assets.BuildRev,
	}
}

func (g Generator) langFor(route kernel.StaticRouteDefinition) string {
	lang := strings.TrimSpace(route.Page.Lang)
	if lang == "" {
		lang = g.assets.DefaultLang
	}

	if lang == "" {
		return "en"
	}

	return lang
}

func (g Generator) titleFor(route kernel.StaticRouteDefinition) string {
	title := strings.TrimSpace(route.Page.Title)
	if title == "" {
		title = g.humanizeRoute(route.Path)
	}

	if g.assets.SiteName == "" {
		return title
	}

	lowerTitle := strings.ToLower(title)
	lowerSite := strings.ToLower(g.assets.SiteName)

	if strings.Contains(lowerTitle, lowerSite) {
		return title
	}

	return fmt.Sprintf("%s | %s", title, g.assets.SiteName)
}

func (g Generator) descriptionFor(route kernel.StaticRouteDefinition) string {
	description := strings.TrimSpace(route.Page.Description)
	if description != "" {
		return description
	}

	base := g.humanizeRoute(route.Path)

	if g.assets.SiteName != "" {
		return fmt.Sprintf("%s page for %s.", base, g.assets.SiteName)
	}

	return fmt.Sprintf("%s page.", base)
}

func (g Generator) canonicalFor(route kernel.StaticRouteDefinition) string {
	canonical := strings.TrimSpace(route.Page.Canonical)
	if canonical != "" {
		return canonical
	}

	if g.assets.CanonicalBase == "" {
		return ""
	}

	path := strings.TrimSpace(route.Path)
	if path == "" {
		path = "/"
	}

	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return g.assets.CanonicalBase + path
}

func (g Generator) filePathFor(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return filepath.Join(g.OutputDir, "index.html")
	}

	parts := strings.Split(trimmed, "/")
	if len(parts) == 1 {
		name := parts[0]
		if name == "" {
			name = "index"
		}

		return filepath.Join(g.OutputDir, name+".html")
	}

	dir := filepath.Join(append([]string{g.OutputDir}, parts[:len(parts)-1]...)...)
	file := parts[len(parts)-1] + ".html"

	return filepath.Join(dir, file)
}

func (g Generator) humanizeRoute(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return "Home"
	}

	cleaned := strings.ReplaceAll(trimmed, "-", " ")
	cleaned = strings.ReplaceAll(cleaned, "_", " ")

	words := strings.Fields(cleaned)
	if len(words) == 0 {
		return "Home"
	}

	for i, word := range words {
		words[i] = capitalize(word)
	}

	return strings.Join(words, " ")
}

func capitalize(word string) string {
	if word == "" {
		return ""
	}

	runes := []rune(strings.ToLower(word))
	if len(runes) == 0 {
		return ""
	}

	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func loadTemplate() (*htmltemplate.Template, error) {
	raw, err := templatesFS.ReadFile(protectedIndexTemplate)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}

	tmpl, err := htmltemplate.New("protected_index").Parse(string(raw))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
