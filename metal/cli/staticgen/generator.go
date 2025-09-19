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

//go:embed templates/public_share.oullin.html
var templatesFS embed.FS

const (
	publicShareTemplate = "templates/public_share.oullin.html"
)

type AssetConfig struct {
	BuildRev      string
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
	Lang           string
	Title          string
	Description    string
	Canonical      string
	Robots         string
	ThemeColor     string
	JsonLD         htmltemplate.JS
	OG             ogData
	Twitter        twitterData
	Hreflangs      []hreflangData
	Favicons       []faviconData
	Manifest       string
	AppleTouchIcon string
	BuildRev       string
}

type ogData struct {
	Type        string
	Image       string
	ImageAlt    string
	ImageWidth  string
	ImageHeight string
	SiteName    string
	Locale      string
}

type twitterData struct {
	Card     string
	Image    string
	ImageAlt string
}

type hreflangData struct {
	Lang string
	Href string
}

type faviconData struct {
	Rel   string
	Href  string
	Type  string
	Sizes string
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
	page := route.Page

	data := templateData{
		Lang:           g.langFor(route),
		Title:          g.titleFor(route),
		Description:    g.descriptionFor(route),
		Canonical:      g.canonicalFor(route),
		Robots:         strings.TrimSpace(page.Robots),
		ThemeColor:     strings.TrimSpace(page.ThemeColor),
		JsonLD:         g.jsonLDFor(page.JsonLD),
		OG:             g.ogFor(route),
		Twitter:        g.twitterFor(page),
		Hreflangs:      g.hreflangsFor(page.Hreflangs),
		Favicons:       g.faviconsFor(page.Favicons),
		Manifest:       strings.TrimSpace(page.Manifest),
		AppleTouchIcon: strings.TrimSpace(page.AppleTouchIcon),
		BuildRev:       g.assets.BuildRev,
	}

	if data.OG.SiteName == "" {
		data.OG.SiteName = g.assets.SiteName
	}

	if data.OG.Locale == "" {
		data.OG.Locale = g.langFor(route)
	}

	return data
}

func (g Generator) jsonLDFor(raw string) htmltemplate.JS {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	return htmltemplate.JS(trimmed)
}

func (g Generator) ogFor(route kernel.StaticRouteDefinition) ogData {
	og := route.Page.OG

	return ogData{
		Type:        strings.TrimSpace(og.Type),
		Image:       strings.TrimSpace(og.Image),
		ImageAlt:    strings.TrimSpace(og.ImageAlt),
		ImageWidth:  strings.TrimSpace(og.ImageWidth),
		ImageHeight: strings.TrimSpace(og.ImageHeight),
		SiteName:    strings.TrimSpace(og.SiteName),
		Locale:      strings.TrimSpace(og.Locale),
	}
}

func (g Generator) twitterFor(page kernel.SharePage) twitterData {
	twitter := page.Twitter

	return twitterData{
		Card:     strings.TrimSpace(twitter.Card),
		Image:    strings.TrimSpace(twitter.Image),
		ImageAlt: strings.TrimSpace(twitter.ImageAlt),
	}
}

func (g Generator) hreflangsFor(items []kernel.Hreflang) []hreflangData {
	if len(items) == 0 {
		return nil
	}

	hreflangs := make([]hreflangData, 0, len(items))
	for _, item := range items {
		lang := strings.TrimSpace(item.Lang)
		href := strings.TrimSpace(item.Href)

		if lang == "" || href == "" {
			continue
		}

		hreflangs = append(hreflangs, hreflangData{Lang: lang, Href: href})
	}

	return hreflangs
}

func (g Generator) faviconsFor(items []kernel.Favicon) []faviconData {
	if len(items) == 0 {
		return nil
	}

	favicons := make([]faviconData, 0, len(items))
	for _, item := range items {
		rel := strings.TrimSpace(item.Rel)
		href := strings.TrimSpace(item.Href)

		if rel == "" || href == "" {
			continue
		}

		favicons = append(favicons, faviconData{
			Rel:   rel,
			Href:  href,
			Type:  strings.TrimSpace(item.Type),
			Sizes: strings.TrimSpace(item.Sizes),
		})
	}

	return favicons
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
	raw, err := templatesFS.ReadFile(publicShareTemplate)
	if err != nil {
		return nil, fmt.Errorf("read template: %w", err)
	}

	tmpl, err := htmltemplate.New("public_share").Parse(string(raw))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
