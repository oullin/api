package seo

import htmltemplate "html/template"

type SEO struct {
	SpaPublicDir string `validate:"required,dirpath"`
	SiteURL      string `validate:"required,url"`
	Lang         string `validate:"required,oneof=en"`
	SiteName     string `validate:"required,min=10"`
}

type TemplateData struct {
	Lang           string          `validate:"required,oneof=en"`
	Title          string          `validate:"required,min=10"`
	Description    string          `validate:"required,min=10"`
	Canonical      string          `validate:"required,url"` //website url
	Robots         string          `validate:"required"`     // default: index,follow
	ThemeColor     string          `validate:"required"`     // default: #0E172B -> dark
	JsonLD         htmltemplate.JS `validate:"required"`
	OGTagOg        TagOgData       `validate:"required"`
	Twitter        TwitterData     `validate:"required"`
	HrefLang       []HrefLangData  `validate:"required"`
	Favicons       []FaviconData   `validate:"required"`
	Manifest       string          `validate:"required"`
	AppleTouchIcon string          `validate:"required"`
}

type TagOgData struct {
	Type        string `validate:"required,oneof=website"` //website
	Image       string `validate:"required,url"`           //https://oullin.io/assets/about-Dt5rMl63.jpg
	ImageAlt    string `validate:"required,min=10"`
	ImageWidth  string `validate:"required"` //600
	ImageHeight string `validate:"required"` //400
	SiteName    string `validate:"required,min=5"`
	Locale      string `validate:"required,min=5"` //en_GB
}

type TwitterData struct {
	Card     string `validate:"required,oneof=summary_large_image"`
	Image    string `validate:"required,url"` //https://oullin.io/assets/about-Dt5rMl63.jpg
	ImageAlt string `validate:"required,min=10"`
}

type HrefLangData struct {
	Lang string `validate:"required,oneof=en"`
	Href string `validate:"required,url"`
}

type FaviconData struct {
	Rel   string `validate:"required,oneof=icon"`
	Href  string `validate:"required,url"`
	Type  string `validate:"required"`
	Sizes string `validate:"required"`
}
