package seo

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

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

	var tData TemplateData
	if tData, err = g.Build(html); err != nil {
		return fmt.Errorf("home: generating template data: %w", err)
	}

	if err = g.Export("index", tData); err != nil {
		return fmt.Errorf("home: exporting template data: %w", err)
	}

	cli.Successln("Home SEO template generated")

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

func (g *Generator) Build(body []template.HTML) (TemplateData, error) {
	og := TagOgData{
		ImageWidth:  "600",
		ImageHeight: "400",
		Type:        "website",
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
		Canonical:      g.Page.SiteURL,
		AppleTouchIcon: g.Page.LogoURL,
		Title:          g.Page.SiteName,
		Categories:     g.Page.Categories,
		JsonLD:         NewJsonID(g.Page).Render(),
		HrefLang: []HrefLangData{
			{Lang: g.Page.Lang, Href: g.Page.SiteURL},
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
