package seo

type Web struct {
	FoundedYear int16
	StubPath    string
	ThemeColor  string
	Robots      string
	ColorScheme string
	Description string
	Pages       map[string]WebPage
}

type WebPage struct {
	Name    string
	Url     string
	Title   string
	Excerpt string
}

func NewWeb() *Web {
	var pages map[string]WebPage

	//const WebHomeUrl = "/"
	//const WebHomeName = "Home"
	home := WebPage{
		Name:    "Home",
		Url:     "/",
		Title:   AuthorName + "'s Personal Website & Journal",
		Excerpt: "Gus's a dedicated engineering leader with over twenty years of experience. He specialises in building high-quality, scalable systems across software development, IT infrastructure, and workplace technology. With expertise in Golang, Node.js, and PHP, He has a proven track record of leading cross-functional teams to deliver secure, compliant solutions, particularly within the financial services sector. His background combines deep technical knowledge in cloud architecture and network protocols with a strategic focus on optimizing workflows, driving innovation, and empowering teams to achieve exceptional results in fast-paced environments.",
	}

	//const WebAboutName = "About"
	//const WebAboutUrl = "/about"
	about := WebPage{
		Name:    "About",
		Url:     "/about",
		Title:   "About " + AuthorName,
		Excerpt: "Gus's an engineering leader who’s passionate about building reliable and smooth software that strive to make a difference. He also has led teams in designing and delivering scalable, high-performance systems that run efficiently even in complex environments",
	}

	//const WebProjectsName = "Projects"
	//const WebProjectsUrl = "/projects"
	projects := WebPage{
		Name:    "Projects",
		Url:     "/projects",
		Title:   AuthorName + "'s Projects",
		Excerpt: "Over the years, Gus’s built and shared command-line tools and frameworks to tackle real engineering challenges—complete with clear docs and automated tests—and partnered with banks, insurers, and fintech to deliver custom software that balances performance, security, and scalability.",
	}

	//const WebResumeName = "Resume"
	//const WebResumeUrl = "/resume"
	resume := WebPage{
		Name:    "Resume",
		Url:     "/resume",
		Title:   AuthorName + "'s Projects",
		Excerpt: "Gus' worked closely with financial services companies, delivering secure and compliant solutions that align with industry regulations and standards. He understands the technical and operational demands of financial institutions and have implemented robust architectures that support high-availability systems, data security, and transactional integrity.",
	}

	//const WebPostsName = "Posts"
	//const WebPostsUrl = "/posts"
	//const WebPostDetailUrl = "/post"
	posts := WebPage{
		Name: "Posts",
		Url:  "/posts",
	}

	postsD := WebPage{
		Name: "Posts",
		Url:  "/post",
	}

	pages[HomeSlug] = home
	pages[AboutSlug] = about
	pages[ProjectsSlug] = projects
	pages[ResumeSlug] = resume
	pages[PostsSlug] = posts
	pages[PostDetailsSlug] = postsD

	return &Web{
		FoundedYear: 2020,
		Pages:       pages,
		StubPath:    "stub.html",
		ThemeColor:  "#0E172B",
		Robots:      "index,follow",
		ColorScheme: "light dark",
		Description: "Gus is a full-stack Software Engineer leader with over two decades of experience in building complex web systems and products, specialising in areas like e-commerce, banking, cross-payment solutions, cyber security, and customer success.",
	}
}

func (w *Web) GetHomePage() WebPage {
	return w.Pages[HomeSlug]
}

func (w *Web) GetAboutPage() WebPage {
	return w.Pages[AboutSlug]
}

func (w *Web) GetResumePage() WebPage {
	return w.Pages[ResumeSlug]
}

func (w *Web) GetProjectsPage() WebPage {
	return w.Pages[ProjectsSlug]
}

func (w *Web) GetPostsPage() WebPage {
	return w.Pages[PostsSlug]
}

func (w *Web) GetPostsDetailPage() WebPage {
	return w.Pages[PostDetailsSlug]
}
