package seo

type Web struct {
	FoundedYear int16
	ThemeColor  string
	Robots      string
	ColorScheme string
	Description string
	Urls        WebPageUrls
	Pages       map[string]WebPage
}

type WebPage struct {
	Name    string
	Url     string
	Title   string
	Excerpt string
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

	home := WebPage{
		Name:    "Home",
		Url:     "/",
		Title:   "Oullin",
		Excerpt: "Oullin is a movement-led platform for engineering leadership, AI architecture, open-source systems, and writing shaped by presence, transformation, and craft.",
	}

	about := WebPage{
		Name:    "About",
		Url:     "/about",
		Title:   "About",
		Excerpt: "Learn how Oullin approaches movement, transformation, and craft, and meet founder Gustavo Ocanto.",
	}

	contact := WebPage{
		Name:    "Contact",
		Url:     "/contact",
		Title:   "Contact",
		Excerpt: "Contact Oullin for engineering leadership, AI architecture, open-source systems, partnerships, and advisory work.",
	}

	projects := WebPage{
		Name:    "Projects",
		Url:     "/projects",
		Title:   "Projects",
		Excerpt: "Explore open-source tools, internal platforms, and client systems from Oullin, built for performance, security, maintainability, and real operating constraints.",
	}

	writing := WebPage{
		Name:    "Writing",
		Url:     "/writing",
		Title:   "Writing Archive",
		Excerpt: "Browse Oullin's writing archive for technical essays, architecture notes, and category-based reading across software, AI, and systems thinking.",
	}

	terms := WebPage{
		Name:    "Terms",
		Url:     "/terms-and-conditions",
		Title:   "Terms and Policies",
		Excerpt: "Review Oullin's terms and policies for software products, consulting, technical architecture, billing, acceptable use, and service responsibilities.",
	}

	postDetail := WebPage{
		// Title and Excerpt are populated per post during page generation.
		Name: "Post",
		Url:  "/post",
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
		Urls:        urls,
		Pages:       pages,
		ThemeColor:  "#0E172B",
		Robots:      "index,follow",
		ColorScheme: "light dark",
		Description: "Oullin is a movement-led platform for engineering leadership, AI architecture, open-source systems, and writing shaped by presence, transformation, and craft.",
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
