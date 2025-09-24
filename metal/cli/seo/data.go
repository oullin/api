package seo

import "html/template"

type Page struct {
	OutputDir     string             `validate:"required"`
	Template      *template.Template `validate:"required"`
	SiteName      string             `validate:"required"`
	SameAsURL     []string           `validate:"required"`
	SiteURL       string             `validate:"required,uri"`
	LogoURL       string             `validate:"required,uri"`
	WebRepoURL    string             `validate:"required,uri"`
	APIRepoURL    string             `validate:"required,uri"`
	AboutPhotoUrl string             `validate:"required,uri"`
	Lang          string             `validate:"required,oneof=en_GB"`
	StubPath      string             `validate:"required,oneof=stub.html"`
}

type TemplateData struct {
	Lang           string          `validate:"required,oneof=en_GB"`
	Title          string          `validate:"required,min=10"`
	Description    string          `validate:"required,min=10"`
	Canonical      string          `validate:"required,url"`
	Robots         string          `validate:"required"`
	ThemeColor     string          `validate:"required"`
	JsonLD         template.JS     `validate:"required"`
	OGTagOg        TagOgData       `validate:"required"`
	Twitter        TwitterData     `validate:"required"`
	HrefLang       []HrefLangData  `validate:"required"`
	Favicons       []FaviconData   `validate:"required"`
	Manifest       template.JS     `validate:"required"`
	AppleTouchIcon string          `validate:"required"`
	Categories     []string        `validate:"required"`
	BgColor        string          `validate:"required"`
	Body           []template.HTML `validate:"required"`
}

type TagOgData struct {
	Type        string `validate:"required,oneof=website"`
	Image       string `validate:"required,url"`
	ImageAlt    string `validate:"required,min=10"`
	ImageWidth  string `validate:"required"`
	ImageHeight string `validate:"required"`
	SiteName    string `validate:"required,min=5"`
	Locale      string `validate:"required,min=5"`
}

type TwitterData struct {
	Card     string `validate:"required,oneof=summary_large_image"`
	Image    string `validate:"required,url"`
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
