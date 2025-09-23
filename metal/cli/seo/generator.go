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
	StubPath  string
	OutputDir string
	Lang      string
	SiteName  string
	SiteURL   string
	HTML      *htmltemplate.Template
}

type Generator struct {
	Tmpl          Template
	Env           *env.Environment
	Validator     *portal.Validator
	DB            *database.Connection
	WebsiteRoutes *router.WebsiteRoutes
	Router        *router.Router //@todo Remove!
}

func NewGenerator(db *database.Connection, env *env.Environment, val *portal.Validator) (*Generator, error) {
	template := Template{
		StubPath:  "stub.html",
		SiteURL:   env.App.URL,
		SiteName:  env.App.Name,
		Lang:      env.App.Lang(),
		OutputDir: env.Seo.SpaDir,
	}

	if html, err := template.LoadTemplate(); err != nil {
		return nil, fmt.Errorf("there was an issue loading the template [%s]: %w", template.StubPath, err)
	} else {
		template.HTML = html
	}

	return &Generator{
		DB:            db,
		Env:           env,
		Tmpl:          template,
		Validator:     val,
		Router:        &router.Router{},
		WebsiteRoutes: router.NewWebsiteRoutes(env),
	}, nil
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

	var data = struct {
		Talks    payload.TalksResponse
		Projects payload.ProjectsResponse
		Profile  payload.ProfileResponse
	}{
		Talks:    talks,
		Profile:  profile,
		Projects: projects,
	}

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
	if err = g.Tmpl.HTML.Execute(&buffer, data); err != nil {
		return fmt.Errorf("home: rendering template: %w", err)
	}

	if err = os.MkdirAll(filepath.Dir(g.Tmpl.OutputDir), 0o755); err != nil {
		return fmt.Errorf("home: creating directory for %s: %w", g.Tmpl.OutputDir, err)
	}

	if err = os.WriteFile(g.Tmpl.OutputDir, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", g.Tmpl.OutputDir, err)
	}

	fmt.Println("Home: Done.")

	return nil
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
