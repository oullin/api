package seo

import (
	"embed"
	"fmt"
	htmltemplate "html/template"

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
	tmpl := Template{
		StubPath:  "stub.html",
		SiteURL:   env.App.URL,
		SiteName:  env.App.Name,
		Lang:      env.App.Lang(),
		OutputDir: env.Seo.SpaDir,
	}

	return &Generator{
		DB:            db,
		Env:           env,
		Tmpl:          tmpl,
		Validator:     val,
		Router:        &router.Router{},
		WebsiteRoutes: router.NewWebsiteRoutes(env),
	}, nil
}

func (g *Generator) GenerateHome() error {
	_, err := g.Tmpl.LoadTemplate()

	if err != nil {
		return fmt.Errorf("loading template: %w", err)
	}

	web := g.WebsiteRoutes
	resource := make(map[string]func() router.StaticRouteResource)

	resource[router.FixtureProfile] = func() router.StaticRouteResource {
		return handler.MakeProfileHandler(web.Fixture.GetProfileFile())
	}

	resource[router.FixtureTalks] = func() router.StaticRouteResource {
		return handler.MakeTalksHandler(web.Fixture.GetTalksFile())
	}

	resource[router.FixtureProjects] = func() router.StaticRouteResource {
		return handler.MakeProjectsHandler(web.Fixture.GetProfileFile())
	}

	var talks payload.TalksResponse
	var profile payload.ProfileResponse
	var projects payload.ProjectsResponse

	if err = Fetch[payload.ProfileResponse](&profile, resource[router.FixtureProfile]); err != nil {
		return fmt.Errorf("error fetching profile: %w", err)
	}

	if err = Fetch[payload.TalksResponse](&talks, resource[router.FixtureTalks]); err != nil {
		return fmt.Errorf("error fetching talks: %w", err)
	}

	if err = Fetch[payload.ProjectsResponse](&projects, resource[router.FixtureProjects]); err != nil {
		return fmt.Errorf("error fetching projects: %w", err)
	}

	fmt.Println("Here: ", profile)

	//PrintResponse(rr)

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
