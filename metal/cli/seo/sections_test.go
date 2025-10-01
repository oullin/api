package seo

import (
	"html/template"
	"strings"
	"testing"
	"time"

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

	social := &payload.SocialResponse{
		Data: []payload.SocialData{{
			Name:        "X",
			Handle:      "@<gocanto>",
			URL:         "https://social.example/<gocanto>",
			Description: "Follow <me>",
		}},
	}

	recommendations := &payload.RecommendationsResponse{
		Data: []payload.RecommendationsData{{
			Relation: "Colleague <and> friend",
			Text:     "Great<br/>lead",
			Person: payload.RecommendationsPersonData{
				FullName:    "Jane <Doe>",
				Company:     "Tech <Corp>",
				Designation: "CTO",
			},
		}},
	}

	experience := &payload.ExperienceResponse{
		Data: []payload.ExperienceData{{
			Company:        "Perx <Tech>",
			Position:       "Head <Engineer>",
			EmploymentType: "Full-Time",
			LocationType:   "Remote",
			City:           "Sing<apore>",
			Country:        "Singa<pore>",
			StartDate:      "2020",
			EndDate:        "2024",
			Summary:        "Led <teams>",
			Skills:         "Go, <PHP>",
		}},
	}

	education := &payload.EducationResponse{
		Data: []payload.EducationData{{
			School:         "Uni<versity>",
			Degree:         "BSc",
			Field:          "Computer <Science>",
			Description:    "Studied <algorithms>",
			GraduatedAt:    "2012",
			IssuingCountry: "Vene<zuela>",
		}},
	}

	categories := []string{"Go<Lang>", "CLI"}

	publishedAt := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	post := &payload.PostResponse{
		Title:   "Building <APIs>",
		Excerpt: "Learn <fast>\nwith examples",
		Content: "Intro paragraph with <tags>\nmore info.\n\nSecond paragraph & details.",
		Author: payload.UserResponse{
			DisplayName: "Gus <C>",
			Username:    "gocanto<script>",
			FirstName:   "Gus<",
			LastName:    "C>",
		},
		Categories:  []payload.CategoryResponse{{Name: "Go<Lang>"}},
		Tags:        []payload.TagResponse{{Name: "SEO<Tag>"}},
		PublishedAt: &publishedAt,
	}

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

	renderedSocial := string(sections.Social(social))
	if strings.Contains(renderedSocial, "<gocanto>") {
		t.Fatalf("social section should escape handles: %q", renderedSocial)
	}

	renderedRecommendations := string(sections.Recommendations(recommendations))
	if !strings.Contains(renderedRecommendations, "Great<br/>lead") {
		t.Fatalf("recommendations should allow br tags: %q", renderedRecommendations)
	}

	renderedExperience := string(sections.Experience(experience))
	if strings.Contains(renderedExperience, "Perx <Tech>") {
		t.Fatalf("experience should escape HTML: %q", renderedExperience)
	}

	renderedEducation := string(sections.Education(education))
	if strings.Contains(renderedEducation, "Uni<versity>") {
		t.Fatalf("education should escape HTML: %q", renderedEducation)
	}

	renderedPost := string(sections.Post(post))
	if strings.Contains(renderedPost, "<APIs>") {
		t.Fatalf("post title should be escaped: %q", renderedPost)
	}
	if strings.Contains(renderedPost, "gocanto<script>") {
		t.Fatalf("post meta should escape username: %q", renderedPost)
	}
	if !strings.Contains(renderedPost, "Learn &lt;fast&gt;<br/>") {
		t.Fatalf("post excerpt should escape html and keep breaks: %q", renderedPost)
	}
	if !strings.Contains(renderedPost, "Second paragraph &amp; details.") {
		t.Fatalf("post content should escape html entities: %q", renderedPost)
	}

	for _, html := range []template.HTML{
		sections.Profile(profile),
		sections.Categories(categories),
		sections.Skills(profile),
		sections.Talks(talks),
		sections.Projects(projects),
		sections.Social(social),
		sections.Recommendations(recommendations),
		sections.Experience(experience),
		sections.Education(education),
		sections.Post(post),
	} {
		if !strings.HasPrefix(string(html), "<h1>") {
			t.Fatalf("section should start with heading: %q", html)
		}
	}
}

func TestSectionsGuardNilInputs(t *testing.T) {
	sections := NewSections()

	if html := sections.Profile(nil); html != template.HTML("") {
		t.Fatalf("expected empty html for nil profile, got %q", html)
	}

	if html := sections.Skills(nil); html != template.HTML("") {
		t.Fatalf("expected empty html for nil skills profile, got %q", html)
	}

	if html := sections.Talks(nil); html != template.HTML("") {
		t.Fatalf("expected empty html for nil talks, got %q", html)
	}

	if html := sections.Post(nil); html != template.HTML("") {
		t.Fatalf("expected empty html for nil post, got %q", html)
	}
}
