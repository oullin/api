package seo

import (
	"html/template"
	"strings"

	"github.com/oullin/handler/payload"
)

type Sections struct{}

func NewSections() Sections {
	return Sections{}
}

func (s *Sections) Profile(profile *payload.ProfileResponse) template.HTML {
	return "<h1>Profile</h1>" +
		template.HTML("<p>"+
			template.HTMLEscapeString(profile.Data.Name)+", "+
			template.HTMLEscapeString(profile.Data.Profession)+
			"</p>",
		)
}

func (s *Sections) Skills(profile *payload.ProfileResponse) template.HTML {
	var items []string

	for _, item := range profile.Data.Skills {
		items = append(items, "<li>"+template.HTMLEscapeString(item.Item)+"</li>")
	}

	return template.HTML("<h1>Skills</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) Talks(talks *payload.TalksResponse) template.HTML {
	var items []string

	for _, item := range talks.Data {
		items = append(items, "<li>"+
			template.HTMLEscapeString(item.Title)+": "+
			template.HTMLEscapeString(item.Subject)+
			"</li>",
		)
	}

	return template.HTML("<h1>Talks</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) Projects(projects *payload.ProjectsResponse) template.HTML {
	var items []string

	for _, item := range projects.Data {
		items = append(items, "<li>"+
			template.HTMLEscapeString(item.Title)+": "+
			template.HTMLEscapeString(item.Excerpt)+
			"</li>",
		)
	}

	return template.HTML("<h1>Projects</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}
