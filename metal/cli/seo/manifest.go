package seo

import (
	"bytes"
	"encoding/json"
	"html/template"
	"time"
)

type Manifest struct {
	Name        string
	ShortName   string
	Description string
	StartURL    string
	Scope       string
	Lang        string
	ThemeColor  string
	BgColor     string
	Categories  []string
	Icons       []ManifestIcon
	Now         func() time.Time
	Shortcuts   []ManifestShortcut
}

type ManifestIcon struct {
	Src     string `json:"src"`
	Type    string `json:"type"`
	Sizes   string `json:"sizes"`
	Purpose string `json:"purpose,omitempty"`
}

type ManifestShortcut struct {
	URL       string         `json:"url"`
	Name      string         `json:"name"`
	ShortName string         `json:"short_name"`
	Icons     []ManifestIcon `json:"icons,omitempty"`
	Desc      string         `json:"description,omitempty"`
}

func NewManifest(tmpl Template, data TemplateData) *Manifest {
	favicon := data.Favicons[1]
	icons := []ManifestIcon{
		{Src: favicon.Href, Sizes: favicon.Sizes, Type: favicon.Type, Purpose: "any"},
	}

	b := &Manifest{
		Icons:       icons,
		Lang:        tmpl.Lang,
		Scope:       WebHomeUrl,
		BgColor:     data.BgColor,
		StartURL:    tmpl.SiteURL,
		Name:        tmpl.SiteName,
		ShortName:   tmpl.SiteName,
		Categories:  data.Categories,
		ThemeColor:  data.ThemeColor,
		Description: data.Description,
		Now:         func() time.Time { return time.Now().UTC() },
		Shortcuts: []ManifestShortcut{
			{
				Icons:     icons,
				URL:       WebHomeUrl,
				Name:      WebHomeName,
				ShortName: WebHomeName,
			},
			{
				Icons:     icons,
				URL:       WebPostUrl,
				Name:      WebPostName,
				ShortName: WebPostName,
			},
			{
				Icons:     icons,
				URL:       WebProjectsUrl,
				Name:      WebProjectsName,
				ShortName: WebProjectsName,
			},
			{
				Icons:     icons,
				URL:       WebAboutUrl,
				Name:      WebAboutName,
				ShortName: WebAboutName,
			},
			{
				Icons:     icons,
				URL:       WebResumeUrl,
				Name:      WebResumeName,
				ShortName: WebResumeName,
			},
		},
	}

	return b
}

func (m *Manifest) Render() template.JS {
	root := map[string]any{
		"dir":                         "ltr",
		"orientation":                 "any",
		"prefer_related_applications": false,
		"name":                        m.Name,
		"lang":                        m.Lang,
		"scope":                       m.Scope,
		"icons":                       m.Icons,
		"screenshots":                 []any{},
		"background_color":            m.BgColor,
		"start_url":                   m.StartURL,
		"short_name":                  m.ShortName,
		"shortcuts":                   m.Shortcuts,
		"display":                     "standalone",
		"theme_color":                 m.ThemeColor,
		"categories":                  m.Categories,
		"description":                 m.Description,
		"id":                          m.StartURL + "#websitemanifest",
	}

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(root); err != nil {
		return `{}`
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, buf.Bytes()); err != nil {
		return template.JS(buf.String())
	}

	return template.JS(compact.String())
}
