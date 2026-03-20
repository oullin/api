package seo

import (
	"fmt"
	"strings"
)

type Brand struct {
	Name string
}

func NewBrand() Brand {
	return Brand{Name: "Oullin"}
}

func (b Brand) TitleFor(pageTitle string) string {
	trimmed := strings.TrimSpace(pageTitle)
	if trimmed == "" || trimmed == b.Name {
		return b.Name
	}

	return fmt.Sprintf("%s - %s", trimmed, b.Name)
}

type Web struct {
	FoundedYear int16
	Brand       Brand
	ThemeColor  string
	Robots      string
	ColorScheme string
	Description string
	Urls        WebPageUrls
	Pages       map[string]WebPage
}

type WebPage struct {
	Name       string
	Url        string
	Title      string
	Excerpt    string
	ImageAlt   string
	SchemaName string
}

type WebPageUrls struct {
	OrganizationURL string
	RepoApiUrl      string
	RepoWebUrl      string
	LogoUrl         string
	AboutPhotoUrl   string
}

func NewWeb() *Web {
	pages := make(map[string]WebPage, 7)
	brand := NewBrand()

	home := WebPage{
		Name:       "Home",
		Url:        "/",
		Title:      brand.Name,
		Excerpt:    "Oullin is a boutique software engineering and architecture consultancy for startups and scale-ups navigating the AI era. AI architecture, modernisation, and resilient systems for regulated and high-trust environments.",
		ImageAlt:   "Oullin brand preview",
		SchemaName: "Oullin",
	}

	about := WebPage{
		Name:       "About",
		Url:        "/about",
		Title:      "About",
		Excerpt:    "About Oullin, a boutique software engineering and architecture consultancy focused on resilient systems, AI-era modernisation, and engineering judgment in regulated and high-trust environments.",
		ImageAlt:   "Oullin brand story",
		SchemaName: "About Oullin",
	}

	contact := WebPage{
		Name:       "Contact",
		Url:        "/contact",
		Title:      "Contact",
		Excerpt:    "Contact Oullin about AI architecture, modernisation, and resilient delivery in regulated and high-trust environments. Direct, senior guidance from the first exchange.",
		ImageAlt:   "Oullin contact page preview",
		SchemaName: "Contact Oullin",
	}

	projects := WebPage{
		Name:       "Projects",
		Url:        "/projects",
		Title:      "Projects",
		Excerpt:    "Proof from real systems: open-source tools, internal platforms, and client work across banking, fintech, AI-era architecture, and resilient software delivery.",
		ImageAlt:   "Oullin project collection preview",
		SchemaName: "Oullin Projects",
	}

	writing := WebPage{
		Name:       "Writing",
		Url:        "/writing",
		Title:      "Writing",
		Excerpt:    "Field notes from real systems: case studies, technical essays, and use cases on AI architecture, production systems, and engineering judgment.",
		ImageAlt:   "Oullin writing archive preview",
		SchemaName: "Oullin Writing",
	}

	terms := WebPage{
		Name:       "Terms",
		Url:        "/terms-and-conditions",
		Title:      "Terms and Policies",
		Excerpt:    "Review Oullin's terms and policies for consulting, technical architecture, software products, billing, acceptable use, and service responsibilities.",
		ImageAlt:   "Oullin terms and policies preview",
		SchemaName: "Oullin Terms and Policies",
	}

	postDetail := WebPage{
		// Title and Excerpt are populated per post during page generation.
		Name:       "Post",
		Url:        "/post",
		ImageAlt:   "Oullin article preview",
		SchemaName: "Oullin Article",
	}

	pages[HomeSlug] = home
	pages[AboutSlug] = about
	pages[ContactSlug] = contact
	pages[ProjectsSlug] = projects
	pages[WritingSlug] = writing
	pages[TermsSlug] = terms
	pages[PostDetailsSlug] = postDetail

	urls := WebPageUrls{
		OrganizationURL: "https://github.com/oullin",
		RepoApiUrl:      "https://github.com/oullin/api",
		RepoWebUrl:      "https://github.com/oullin/web",
		LogoUrl:         "https://oullin.io/assets/001-BBig3EFt.png",
		AboutPhotoUrl:   "https://oullin.io/images/profile/about-seo.png",
	}

	return &Web{
		FoundedYear: 2020,
		Brand:       brand,
		Urls:        urls,
		Pages:       pages,
		ThemeColor:  "#0E172B",
		Robots:      "index,follow",
		ColorScheme: "light dark",
		Description: "Oullin is a boutique software engineering and architecture consultancy for startups and scale-ups navigating the AI era. AI architecture, modernisation, and resilient systems for regulated and high-trust environments.",
	}
}

func (w *Web) GetHomePage() WebPage {
	return w.Pages[HomeSlug]
}

func (w *Web) GetAboutPage() WebPage {
	return w.Pages[AboutSlug]
}

func (w *Web) GetContactPage() WebPage {
	return w.Pages[ContactSlug]
}

func (w *Web) GetProjectsPage() WebPage {
	return w.Pages[ProjectsSlug]
}

func (w *Web) GetWritingPage() WebPage {
	return w.Pages[WritingSlug]
}

func (w *Web) GetTermsPage() WebPage {
	return w.Pages[TermsSlug]
}

func (w *Web) GetPostDetailPage() WebPage {
	return w.Pages[PostDetailsSlug]
}
