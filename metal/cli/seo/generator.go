package seo

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/oullin/database"
	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/portal"
)

//go:embed stub.html
var templatesFS embed.FS

type Generator struct {
	Page          Page
	Client        *Client
	Env           *env.Environment
	Validator     *portal.Validator
	DB            *database.Connection
	WebsiteRoutes *router.WebsiteRoutes
}

func NewGenerator(db *database.Connection, env *env.Environment, val *portal.Validator) (*Generator, error) {
	var err error
	var categories []string
	var html *template.Template

	if categories, err = NewCategories(db).Generate(); err != nil {
		return nil, err
	}

	page := Page{
		LogoURL:       LogoUrl,
		StubPath:      StubPath,
		WebRepoURL:    RepoWebUrl,
		APIRepoURL:    RepoApiUrl,
		Categories:    categories,
		SiteURL:       env.App.URL,
		SiteName:      env.App.Name,
		AboutPhotoUrl: AboutPhotoUrl,
		Lang:          env.App.Lang(),
		OutputDir:     env.Seo.SpaDir,
		Template:      &template.Template{},
		SameAsURL:     []string{RepoApiUrl, RepoWebUrl, GocantoUrl},
	}

	if _, err = val.Rejects(page); err != nil {
		return nil, fmt.Errorf("invalid template state: %s", val.GetErrorsAsJson())
	}

	if html, err = page.Load(); err != nil {
		return nil, fmt.Errorf("could not load initial stub: %w", err)
	} else {
		page.Template = html
	}

	webRoutes := router.NewWebsiteRoutes(env)

	return &Generator{
		DB:            db,
		Env:           env,
		Validator:     val,
		Page:          page,
		WebsiteRoutes: webRoutes,
		Client:        NewClient(webRoutes),
	}, nil
}

func (g *Generator) Generate() error {
	var err error

	if err = g.GenerateIndex(); err != nil {
		return err
	}

	if err = g.GenerateAbout(); err != nil {
		return err
	}

	if err = g.GenerateProjects(); err != nil {
		return err
	}

	if err = g.GenerateResume(); err != nil {
		return err
	}

	return nil
}

func (g *Generator) GenerateIndex() error {
	var err error
	var talks *payload.TalksResponse
	var profile *payload.ProfileResponse
	var projects *payload.ProjectsResponse

	if profile, err = g.Client.GetProfile(); err != nil {
		return err
	}

	if talks, err = g.Client.GetTalks(); err != nil {
		return err
	}

	if projects, err = g.Client.GetProjects(); err != nil {
		return err
	}

	var html []template.HTML
	sections := NewSections()

	html = append(html, sections.Profile(profile))
	html = append(html, sections.Categories(g.Page.Categories))
	html = append(html, sections.Talks(talks))
	html = append(html, sections.Skills(profile))
	html = append(html, sections.Projects(projects))

	// ----- Template Parsing

	tData, buildErr := g.buildForPage(WebHomeName, WebHomeUrl, html)
	if buildErr != nil {
		return fmt.Errorf("home: generating template data: %w", buildErr)
	}

	if err = g.Export("index", tData); err != nil {
		return fmt.Errorf("home: exporting template data: %w", err)
	}

	cli.Successln("Home SEO template generated")

	return nil
}

func (g *Generator) GenerateAbout() error {
	profile, err := g.Client.GetProfile()
	if err != nil {
		return err
	}

	social, err := g.Client.GetSocial()
	if err != nil {
		return err
	}

	recommendations, err := g.Client.GetRecommendations()
	if err != nil {
		return err
	}

	sections := NewSections()
	var html []template.HTML

	html = append(html, sections.Profile(profile))
	html = append(html, sections.Social(social))
	html = append(html, sections.Recommendations(recommendations))

	data, buildErr := g.buildForPage(WebAboutName, WebAboutUrl, html)
	if buildErr != nil {
		return fmt.Errorf("about: generating template data: %w", buildErr)
	}

	if err = g.Export("about", data); err != nil {
		return fmt.Errorf("about: exporting template data: %w", err)
	}

	cli.Successln("About SEO template generated")

	return nil
}

func (g *Generator) GenerateProjects() error {
	projects, err := g.Client.GetProjects()
	if err != nil {
		return err
	}

	sections := NewSections()
	body := []template.HTML{sections.Projects(projects)}

	data, buildErr := g.buildForPage(WebProjectsName, WebProjectsUrl, body)
	if buildErr != nil {
		return fmt.Errorf("projects: generating template data: %w", buildErr)
	}

	if err = g.Export("projects", data); err != nil {
		return fmt.Errorf("projects: exporting template data: %w", err)
	}

	cli.Successln("Projects SEO template generated")

	return nil
}

func (g *Generator) GenerateResume() error {
	experience, err := g.Client.GetExperience()
	if err != nil {
		return err
	}

	education, err := g.Client.GetEducation()
	if err != nil {
		return err
	}

	recommendations, err := g.Client.GetRecommendations()
	if err != nil {
		return err
	}

	sections := NewSections()
	var html []template.HTML

	html = append(html, sections.Experience(experience))
	html = append(html, sections.Education(education))
	html = append(html, sections.Recommendations(recommendations))

	data, buildErr := g.buildForPage(WebResumeName, WebResumeUrl, html)
	if buildErr != nil {
		return fmt.Errorf("resume: generating template data: %w", buildErr)
	}

	if err = g.Export("resume", data); err != nil {
		return fmt.Errorf("resume: exporting template data: %w", err)
	}

	cli.Successln("Resume SEO template generated")

	return nil
}

func (g *Generator) Export(origin string, data TemplateData) error {
	var err error
	var buffer bytes.Buffer
	fileName := fmt.Sprintf("%s.seo.html", origin)

	cli.Warningln("Executing file: " + fileName)
	if err = g.Page.Template.Execute(&buffer, data); err != nil {
		return fmt.Errorf("%s: rendering template: %w", fileName, err)
	}

	cli.Cyanln(fmt.Sprintf("Working on directory: %s", g.Page.OutputDir))
	if err = os.MkdirAll(g.Page.OutputDir, 0o755); err != nil {
		return fmt.Errorf("%s: creating directory for %s: %w", fileName, g.Page.OutputDir, err)
	}

	out := filepath.Join(g.Page.OutputDir, fileName)
	cli.Blueln(fmt.Sprintf("Writing file on: %s", out))
	if err = os.WriteFile(out, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("%s: writing %s: %w", fileName, out, err)
	}

	cli.Grayln(fmt.Sprintf("File %s generated at: %s", fileName, out))
	cli.Grayln("------------------")

	return nil
}

func (g *Generator) buildForPage(pageName, path string, body []template.HTML) (TemplateData, error) {
	og := TagOgData{
		ImageHeight: "630",
		ImageWidth:  "1200",
		Type:        "website",
		ImageType:   "image/png",
		Locale:      g.Page.Lang,
		ImageAlt:    g.Page.SiteName,
		SiteName:    g.Page.SiteName,
		Image:       g.Page.AboutPhotoUrl,
	}

	twitter := TwitterData{
		Card:     "summary_large_image",
		Image:    g.Page.AboutPhotoUrl,
		ImageAlt: g.Page.SiteName,
	}

	data := TemplateData{
		OGTagOg:        og,
		Robots:         Robots,
		Twitter:        twitter,
		ThemeColor:     ThemeColor,
		ColorScheme:    ColorScheme,
		BgColor:        ThemeColor,
		Lang:           g.Page.Lang,
		Description:    Description,
		AppleTouchIcon: g.Page.LogoURL,
		Categories:     g.Page.Categories,
		JsonLD:         NewJsonID(g.Page).Render(),
		HrefLang: []HrefLangData{
			{Lang: g.Page.Lang, Href: g.canonicalFor(path)},
		},
		Favicons: []FaviconData{
			{
				Rel:   "icon",
				Sizes: "48x48",
				Type:  "image/ico",
				Href:  g.Page.SiteURL + "/favicon.ico",
			},
		},
	}

	data.Body = body
	data.Canonical = g.canonicalFor(path)
	data.Title = g.titleFor(pageName)
	data.Manifest = NewManifest(g.Page, data).Render()

	if _, err := g.Validator.Rejects(og); err != nil {
		return TemplateData{}, fmt.Errorf("invalid og data: %s", g.Validator.GetErrorsAsJson())
	}

	if _, err := g.Validator.Rejects(twitter); err != nil {
		return TemplateData{}, fmt.Errorf("invalid twitter data: %s", g.Validator.GetErrorsAsJson())
	}

	if _, err := g.Validator.Rejects(data); err != nil {
		return TemplateData{}, fmt.Errorf("invalid template data: %s", g.Validator.GetErrorsAsJson())
	}

	return data, nil
}

func (t *Page) Load() (*template.Template, error) {
	raw, err := templatesFS.ReadFile(t.StubPath)
	if err != nil {
		return nil, fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := template.
		New("seo").
		Funcs(template.FuncMap{
			"ManifestDataURL": ManifestDataURL,
		}).
		Parse(string(raw))

	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return tmpl, nil
}

func (g *Generator) canonicalFor(path string) string {
	base := strings.TrimSuffix(g.Page.SiteURL, "/")

	if path == "" || path == "/" {
		return base
	}

	if strings.HasSuffix(base, path) {
		return base
	}

	return base + path
}

func (g *Generator) titleFor(pageName string) string {
	if pageName == WebHomeName {
		return g.Page.SiteName
	}

	return fmt.Sprintf("%s · %s", pageName, g.Page.SiteName)
}
