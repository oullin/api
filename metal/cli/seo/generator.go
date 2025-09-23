package seo

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"os"
	"path/filepath"

	"github.com/oullin/database"
	"github.com/oullin/handler"
	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/portal"
)

//go:embed stub.html
var templatesFS embed.FS

type Template struct {
	StubPath      string                 `validate:"required,oneof=stub.html"`
	OutputDir     string                 `validate:"required"`
	Lang          string                 `validate:"required,oneof=en_GB"`
	SiteName      string                 `validate:"required"`
	SiteURL       string                 `validate:"required,uri"`
	LogoURL       string                 `validate:"required,uri"`
	SameAsURL     []string               `validate:"required"`
	WebRepoURL    string                 `validate:"required,uri"`
	APIRepoURL    string                 `validate:"required,uri"`
	AboutPhotoUrl string                 `validate:"required,uri"`
	HTML          *htmltemplate.Template `validate:"required"`
}

type Generator struct {
	Tmpl          Template              `validate:"required"`
	Env           *env.Environment      `validate:"required"`
	Validator     *portal.Validator     `validate:"required"`
	DB            *database.Connection  `validate:"required"`
	WebsiteRoutes *router.WebsiteRoutes `validate:"required"`
}

func NewGenerator(db *database.Connection, env *env.Environment, val *portal.Validator) (*Generator, error) {
	template := Template{
		LogoURL:       LogoUrl,
		WebRepoURL:    RepoWebUrl,
		APIRepoURL:    RepoApiUrl,
		StubPath:      "stub.html",
		SiteURL:       env.App.URL,
		SiteName:      env.App.Name,
		AboutPhotoUrl: AboutPhotoUrl,
		Lang:          env.App.Lang(),
		OutputDir:     env.Seo.SpaDir,
		HTML:          &htmltemplate.Template{},
		SameAsURL:     []string{RepoApiUrl, RepoWebUrl, GocantoUrl},
	}

	if html, err := template.LoadTemplate(); err != nil {
		return nil, fmt.Errorf("there was an issue loading the template [%s]: %w", template.StubPath, err)
	} else {
		template.HTML = html
	}

	if _, err := val.Rejects(template); err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	gen := Generator{
		DB:            db,
		Env:           env,
		Validator:     val,
		Tmpl:          template,
		WebsiteRoutes: router.NewWebsiteRoutes(env),
	}

	if _, err := val.Rejects(gen); err != nil {
		return nil, fmt.Errorf("invalid generator: %w", err)
	}

	return &gen, nil
}

func (g *Generator) GenerateHome() error {
	var err error
	web := g.WebsiteRoutes
	resource := make(map[string]func() router.StaticRouteResource)

	resource[router.FixtureProfile] = func() router.StaticRouteResource {
		return handler.MakeProfileHandler(web.Fixture.GetProfileFile())
	}

	resource[router.FixtureTalks] = func() router.StaticRouteResource {
		return handler.MakeTalksHandler(web.Fixture.GetTalksFile())
	}

	resource[router.FixtureProjects] = func() router.StaticRouteResource {
		return handler.MakeProjectsHandler(web.Fixture.GetProjectsFile())
	}

	var talks payload.TalksResponse
	var profile payload.ProfileResponse
	var projects payload.ProjectsResponse

	if err = Fetch[payload.ProfileResponse](&profile, resource[router.FixtureProfile]); err != nil {
		return fmt.Errorf("home: error fetching profile: %w", err)
	}

	if err = Fetch[payload.TalksResponse](&talks, resource[router.FixtureTalks]); err != nil {
		return fmt.Errorf("home: error fetching talks: %w", err)
	}

	if err = Fetch[payload.ProjectsResponse](&projects, resource[router.FixtureProjects]); err != nil {
		return fmt.Errorf("home: error fetching projects: %w", err)
	}

	var buffer bytes.Buffer
	var data = struct {
		Talks    payload.TalksResponse
		Profile  payload.ProfileResponse
		Projects payload.ProjectsResponse
	}{
		Talks:    talks,
		Profile:  profile,
		Projects: projects,
	}

	if err = g.Tmpl.HTML.Execute(&buffer, data); err != nil {
		return fmt.Errorf("home: rendering template: %w", err)
	}

	if err = os.MkdirAll(g.Tmpl.OutputDir, 0o755); err != nil {
		return fmt.Errorf("home: creating directory for %s: %w", g.Tmpl.OutputDir, err)
	}

	out := filepath.Join(g.Tmpl.OutputDir, "index.html")
	if err = os.WriteFile(out, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", out, err)
	}

	fmt.Println("Home: Done.")

	return nil
}

func (g *Generator) Generate() (TemplateData, error) {
	og := TagOgData{
		ImageWidth:  "600",
		ImageHeight: "400",
		Type:        "website",
		Locale:      g.Tmpl.Lang,
		ImageAlt:    g.Tmpl.SiteName,
		SiteName:    g.Tmpl.SiteName,
		Image:       g.Tmpl.AboutPhotoUrl,
	}

	twitter := TwitterData{
		Card:     "summary_large_image",
		Image:    g.Tmpl.AboutPhotoUrl,
		ImageAlt: g.Tmpl.SiteName,
	}

	data := TemplateData{
		OGTagOg:        og,
		Robots:         Robots,
		Twitter:        twitter,
		ThemeColor:     ThemeColor,
		Lang:           g.Tmpl.Lang,
		Description:    Description,
		Canonical:      g.Tmpl.SiteURL,
		AppleTouchIcon: g.Tmpl.LogoURL,
		Title:          g.Tmpl.SiteName,
		JsonLD:         NewJsonID(g.Tmpl).Render(),
		Categories:     []string{"one", "two"}, //@todo Fetch this!
		HrefLang: []HrefLangData{
			{Lang: g.Tmpl.Lang, Href: g.Tmpl.SiteURL},
		},
		Favicons: []FaviconData{
			{
				Rel:   "icon",
				Sizes: "48x48",
				Type:  "image/ico",
				Href:  g.Tmpl.SiteURL + "/favicon.ico",
			},
		},
	}

	manifest := NewManifest(g.Tmpl, data)

	data.Manifest = manifest.Render()

	if _, err := g.Validator.Rejects(og); err != nil {
		return TemplateData{}, fmt.Errorf("generate: invalid og data: %w", err)
	}

	if _, err := g.Validator.Rejects(twitter); err != nil {
		return TemplateData{}, fmt.Errorf("generate: invalid twitter data: %w", err)
	}

	if _, err := g.Validator.Rejects(twitter); err != nil {
		return TemplateData{}, fmt.Errorf("generate: invalid twitter data: %w", err)
	}

	if _, err := g.Validator.Rejects(data); err != nil {
		return TemplateData{}, fmt.Errorf("generate: invalid template data: %w", err)
	}

	return data, nil
}

func (t *Template) LoadTemplate() (*htmltemplate.Template, error) {
	raw, err := templatesFS.ReadFile(t.StubPath)
	if err != nil {
		return nil, fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := htmltemplate.New("seo").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return tmpl, nil
}
