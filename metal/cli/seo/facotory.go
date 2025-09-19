package seo

import htmltemplate "html/template"

type AssetConfig struct {
	CanonicalBase string
	DefaultLang   string
	SiteName      string
}

type Generator struct {
	OutputDir string
	assets    AssetConfig
	tmpl      *htmltemplate.Template
}

type Handler struct {
	Generator *Generator
}
