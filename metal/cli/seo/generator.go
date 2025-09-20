package seo

import (
	"embed"
	"fmt"
	htmltemplate "html/template"

	"github.com/oullin/database"
	"github.com/oullin/metal/env"
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
	Tmpl Template
	Env  *env.Environment
	DB   *database.Connection
}

func NewGenerator(db *database.Connection, env *env.Environment) (*Generator, error) {
	tmpl := Template{
		StubPath:  "stub.html",
		OutputDir: env.Seo.SpaDir,
		Lang:      env.App.Lang(),
		SiteName:  env.App.Name,
		SiteURL:   env.App.URL,
	}

	return &Generator{
		Tmpl: tmpl,
		Env:  env,
		DB:   db,
	}, nil
}

func (t *Template) LoadTemplate() (*htmltemplate.Template, error) {
	raw, err := templatesFS.ReadFile(t.StubPath)
	if err != nil {
		return nil, fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := htmltemplate.New("public").Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	return tmpl, nil
}
