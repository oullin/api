package seo

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	_ "runtime/cgo"

	"github.com/oullin/database"
	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/router"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/portal"
)

//go:embed stub.html
var templatesFS embed.FS

const cgoEnabled = true

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

	cli.Magentaln("Initialising SEO generator")
	cli.Magentaln(fmt.Sprintf("CGO enabled: %t", cgoEnabled))

	if categories, err = NewCategories(db).Generate(); err != nil {
		return nil, fmt.Errorf("initialising categories: %w", err)
	}

	page := Page{
		StubPath:      StubPath,
		Categories:    categories,
		SiteName:      env.App.Name,
		Lang:          env.App.Lang(),
		OutputDir:     env.Seo.SpaDir,
		Template:      &template.Template{},
		LogoURL:       portal.SanitiseURL(LogoUrl),
		WebRepoURL:    portal.SanitiseURL(RepoWebUrl),
		APIRepoURL:    portal.SanitiseURL(RepoApiUrl),
		SiteURL:       portal.SanitiseURL(env.App.URL),
		AboutPhotoUrl: portal.SanitiseURL(AboutPhotoUrl),
		SameAsURL: []string{
			portal.SanitiseURL(RepoApiUrl),
			portal.SanitiseURL(RepoWebUrl),
			portal.SanitiseURL(GocantoUrl),
		},
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
	cli.Magentaln("Starting SEO generation pipeline")
	defer cli.Magentaln("SEO generation pipeline finished")

	steps := []struct {
		name string
		fn   func() error
	}{
		{"index", g.GenerateIndex},
		{"about", g.GenerateAbout},
		{"projects", g.GenerateProjects},
		{"resume", g.GenerateResume},
		{"posts", g.GeneratePosts},
	}

	for _, step := range steps {
		if err := step.fn(); err != nil {
			return fmt.Errorf("generating %s: %w", step.name, err)
		}
	}

	return nil
}

func (g *Generator) GenerateIndex() error {
	var err error
	var talks *payload.TalksResponse
	var profile *payload.ProfileResponse
	var projects *payload.ProjectsResponse

	cli.Cyanln("Fetching profile for home page")
	if profile, err = g.Client.GetProfile(); err != nil {
		return fmt.Errorf("home: fetching profile: %w", err)
	}

	cli.Cyanln("Fetching talks for home page")
	if talks, err = g.Client.GetTalks(); err != nil {
		return fmt.Errorf("home: fetching talks: %w", err)
	}

	cli.Cyanln("Fetching projects for home page")
	if projects, err = g.Client.GetProjects(); err != nil {
		return fmt.Errorf("home: fetching projects: %w", err)
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
	cli.Cyanln("Fetching profile for about page")
	profile, err := g.Client.GetProfile()
	if err != nil {
		return fmt.Errorf("about: fetching profile: %w", err)
	}

	cli.Cyanln("Fetching social links for about page")
	social, err := g.Client.GetSocial()
	if err != nil {
		return fmt.Errorf("about: fetching social links: %w", err)
	}

	cli.Cyanln("Fetching recommendations for about page")
	recommendations, err := g.Client.GetRecommendations()
	if err != nil {
		return fmt.Errorf("about: fetching recommendations: %w", err)
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
	cli.Cyanln("Fetching projects for projects page")
	projects, err := g.Client.GetProjects()

	if err != nil {
		return fmt.Errorf("projects: fetching projects: %w", err)
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
	cli.Cyanln("Fetching experience for resume page")
	experience, err := g.Client.GetExperience()

	if err != nil {
		return fmt.Errorf("resume: fetching experience: %w", err)
	}

	cli.Cyanln("Fetching education for resume page")
	education, err := g.Client.GetEducation()
	if err != nil {
		return fmt.Errorf("resume: fetching education: %w", err)
	}

	cli.Cyanln("Fetching recommendations for resume page")
	recommendations, err := g.Client.GetRecommendations()
	if err != nil {
		return fmt.Errorf("resume: fetching recommendations: %w", err)
	}

	sections := NewSections()
	var html []template.HTML

	html = append(html, sections.Education(education))
	html = append(html, sections.Experience(experience))
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

func (g *Generator) GeneratePosts() error {
	var posts []database.Post

	err := g.DB.Sql().
		Model(&database.Post{}).
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Where("posts.published_at IS NOT NULL").
		Where("posts.deleted_at IS NULL").
		Order("posts.published_at DESC").
		Find(&posts).Error

	if err != nil {
		return fmt.Errorf("posts: fetching published posts: %w", err)
	}

	if len(posts) == 0 {
		cli.Grayln("No published posts available for SEO generation")
		return nil
	}

	cli.Cyanln(fmt.Sprintf("Preparing SEO for %d posts", len(posts)))

	sections := NewSections()

	for _, post := range posts {
		cli.Cyanln(fmt.Sprintf("Building SEO for post: %s", post.Slug))
		response := payload.GetPostsResponse(post)
		cli.Grayln(fmt.Sprintf("Post slug: %s", response.Slug))
		cli.Grayln(fmt.Sprintf("Post title: %s", response.Title))
		body := []template.HTML{sections.Post(&response)}

		data, buildErr := g.BuildForPost(response, body)
		if buildErr != nil {
			return fmt.Errorf("posts: building seo for %s: %w", response.Slug, buildErr)
		}

		origin := filepath.Join("posts", response.Slug)
		if err = g.Export(origin, data); err != nil {
			return fmt.Errorf("posts: exporting %s: %w", response.Slug, err)
		}

		cli.Successln(fmt.Sprintf("Post SEO template generated for %s", response.Slug))
	}

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

	out := filepath.Join(g.Page.OutputDir, fileName)

	cli.Cyanln(fmt.Sprintf("Working on directory: %s", filepath.Dir(out)))
	if err = os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return fmt.Errorf("%s: creating directory for %s: %w", fileName, filepath.Dir(out), err)
	}

	if err = os.Remove(out); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s: removing stale export %s: %w", fileName, out, err)
	}

	cli.Blueln(fmt.Sprintf("Writing file on: %s", out))
	if err = os.WriteFile(out, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("%s: writing %s: %w", fileName, out, err)
	}

	cli.Grayln(fmt.Sprintf("File %s generated at: %s", fileName, out))
	cli.Grayln("------------------")

	return nil
}

func (g *Generator) buildForPage(pageName, path string, body []template.HTML, opts ...func(*TemplateData)) (TemplateData, error) {
	og := TagOgData{
		ImageHeight: "630",
		ImageWidth:  "1200",
		Type:        "website",
		ImageType:   "image/png",
		Locale:      g.Page.Lang,
		ImageAlt:    g.Page.SiteName,
		SiteName:    g.Page.SiteName,
		Image:       portal.SanitiseURL(g.Page.AboutPhotoUrl),
	}

	twitter := TwitterData{
		ImageAlt: g.Page.SiteName,
		Card:     "summary_large_image",
		Image:    portal.SanitiseURL(g.Page.AboutPhotoUrl),
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
		Categories:     g.Page.Categories,
		JsonLD:         NewJsonID(g.Page).Render(),
		AppleTouchIcon: portal.SanitiseURL(g.Page.LogoURL),
		HrefLang: []HrefLangData{
			{
				Lang: g.Page.Lang,
				Href: portal.SanitiseURL(g.CanonicalFor(path)),
			},
		},
		Favicons: []FaviconData{
			{
				Rel:   "icon",
				Sizes: "48x48",
				Type:  "image/ico",
				Href:  portal.SanitiseURL(g.Page.SiteURL + "/favicon.ico"),
			},
		},
	}

	data.Body = body
	data.Title = g.TitleFor(pageName)
	data.Manifest = NewManifest(g.Page, data).Render()
	data.Canonical = portal.SanitiseURL(g.CanonicalFor(path))

	for _, opt := range opts {
		opt(&data)
	}

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

func (g *Generator) CanonicalFor(path string) string {
	base := strings.TrimSuffix(g.Page.SiteURL, "/")

	if path == "" || path == "/" {
		return base
	}

	if strings.HasSuffix(base, path) {
		return base
	}

	return base + path
}

func (g *Generator) TitleFor(pageName string) string {
	if pageName == WebHomeName {
		return g.Page.SiteName
	}

	return fmt.Sprintf("%s · %s", pageName, g.Page.SiteName)
}

func truncateForLog(value string) string {
	const maxRunes = 80

	cleaned := strings.TrimSpace(value)
	if cleaned == "" {
		return "(empty)"
	}

	if utf8.RuneCountInString(cleaned) <= maxRunes {
		return cleaned
	}

	runes := []rune(cleaned)
	return string(runes[:maxRunes-3]) + "..."
}

func (g *Generator) BuildForPost(post payload.PostResponse, body []template.HTML) (TemplateData, error) {
	path := g.CanonicalPostPath(post.Slug)
	imageAlt := g.SanitizeAltText(post.Title, g.Page.SiteName)
	description := g.SanitizeMetaDescription(post.Excerpt, Description)
	image := g.PreferredImageURL(post.CoverImageURL, g.Page.AboutPhotoUrl)
	imageType := "image/png"

	cli.Grayln(fmt.Sprintf("Preparing post metadata"))
	cli.Grayln(fmt.Sprintf("  Canonical path: %s", path))
	cli.Grayln(fmt.Sprintf("  Sanitised alt text: %s", imageAlt))
	cli.Grayln(fmt.Sprintf("  Description preview: %s", truncateForLog(description)))
	cli.Grayln(fmt.Sprintf("  Preferred image candidate: %s", image))

	if prepared, err := g.preparePostImage(post); err == nil && prepared.URL != "" {
		cli.Grayln(fmt.Sprintf("  Post image prepared at: %s (%s)", prepared.URL, prepared.Mime))
		image = prepared.URL
		imageType = prepared.Mime
	} else if err != nil {
		cli.Errorln(fmt.Sprintf("failed to prepare post image for %s: %v", post.Slug, err))
		cli.Grayln(fmt.Sprintf("  Falling back to preferred image URL: %s", image))
	}

	return g.buildForPage(post.Title, path, body, func(data *TemplateData) {
		data.OGTagOg.Image = image
		data.OGTagOg.Type = "article"
		data.OGTagOg.ImageType = imageType
		data.Twitter.Image = image
		data.Description = description
		data.OGTagOg.ImageAlt = imageAlt
		data.Twitter.ImageAlt = imageAlt
	})
}

func (g *Generator) CanonicalPostPath(slug string) string {
	cleaned := strings.TrimSpace(slug)
	cleaned = strings.Trim(cleaned, "/")

	if cleaned == "" {
		return WebPostDetailUrl
	}

	return WebPostDetailUrl + "/" + cleaned
}

func (g *Generator) SanitizeMetaDescription(raw, fallback string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(raw, "\n", " "))
	if trimmed == "" {
		return fallback
	}

	condensed := strings.Join(strings.Fields(trimmed), " ")
	escaped := template.HTMLEscapeString(condensed)

	if utf8.RuneCountInString(escaped) < 10 {
		return fallback
	}

	return escaped
}

func (g *Generator) PreferredImageURL(candidate, fallback string) string {
	candidate = strings.TrimSpace(candidate)
	if candidate == "" {
		return fallback
	}

	parsed, err := url.ParseRequestURI(candidate)
	if err != nil {
		return fallback
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fallback
	}

	return candidate
}

func (g *Generator) SanitizeAltText(title, site string) string {
	base := strings.TrimSpace(title)

	if base == "" {
		base = site
	}

	alt := strings.Join(strings.Fields(base+" cover image"), " ")
	escaped := template.HTMLEscapeString(alt)

	if utf8.RuneCountInString(escaped) < 10 {
		fallback := template.HTMLEscapeString(site + " cover image")
		if utf8.RuneCountInString(fallback) < 10 {
			return "SEO cover image"
		}

		return fallback
	}

	return escaped
}
