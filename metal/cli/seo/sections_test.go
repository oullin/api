package seo

import (
	"html/template"
	"strings"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestSectionsRenderersEscapeContent(t *testing.T) {
	sections := NewSections()

	profile := &payload.ProfileResponse{
		Data: payload.ProfileDataResponse{
			Name:       "<Gus>",
			Profession: "Engineer & Co<Founder>",
			Skills: []payload.ProfileSkillsResponse{
				{Item: "Go<lang>"},
				{Item: "Vue.js"},
			},
		},
	}

	talks := &payload.TalksResponse{
		Data: []payload.TalksData{{Title: "Intro<Go>", Subject: "Security & Go"}},
	}

	projects := &payload.ProjectsResponse{
		Data: []payload.ProjectsData{{Title: "API<Server>", Excerpt: "CLI & Tools"}},
	}

	categories := []string{"Go<Lang>", "CLI"}

	renderedProfile := string(sections.Profile(profile))
	if strings.Contains(renderedProfile, "<Gus>") {
		t.Fatalf("profile output was not escaped: %q", renderedProfile)
	}

	renderedCategories := string(sections.Categories(categories))
	if strings.Count(renderedCategories, "<li>") != 2 {
		t.Fatalf("expected two categories list items: %q", renderedCategories)
	}

	renderedSkills := string(sections.Skills(profile))
	if !strings.Contains(renderedSkills, "Go&lt;lang&gt;") {
		t.Fatalf("skills should escape html: %q", renderedSkills)
	}

	renderedTalks := string(sections.Talks(talks))
	if strings.Contains(renderedTalks, "Intro<Go>") {
		t.Fatalf("talks output was not escaped: %q", renderedTalks)
	}

	renderedProjects := string(sections.Projects(projects))
	if !strings.Contains(renderedProjects, "API&lt;Server&gt;") {
		t.Fatalf("projects output missing escaped title: %q", renderedProjects)
	}

	for _, html := range []template.HTML{
		sections.Profile(profile),
		sections.Categories(categories),
		sections.Skills(profile),
		sections.Talks(talks),
		sections.Projects(projects),
	} {
		if !strings.HasPrefix(string(html), "<h1>") {
			t.Fatalf("section should start with heading: %q", html)
		}
	}
}
