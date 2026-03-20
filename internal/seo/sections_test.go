package seo_test

import (
	"html/template"
	"strings"
	"testing"
	"time"

	categoriespkg "github.com/oullin/internal/categories"
	"github.com/oullin/internal/education"
	"github.com/oullin/internal/experience"
	"github.com/oullin/internal/links"
	"github.com/oullin/internal/posts"
	"github.com/oullin/internal/profile"
	"github.com/oullin/internal/projects"
	"github.com/oullin/internal/recommendations"
	"github.com/oullin/internal/seo"
	"github.com/oullin/internal/talks"
)

func TestSectionsRenderersEscapeContent(t *testing.T) {
	sections := seo.NewSections()

	profileResponse := &profile.ProfileResponse{
		Data: profile.ProfileDataResponse{
			Name:       "<Gus>",
			Profession: "Engineer & Co<Founder>",
			Skills: []profile.ProfileSkillsResponse{
				{Item: "Go<lang>"},
				{Item: "Vue.js"},
			},
		},
	}

	talksResponse := &talks.TalksResponse{
		Data: []talks.TalksData{{Title: "Intro<Go>", Subject: "Security & Go"}},
	}

	projectsResponse := &projects.ProjectsResponse{
		Data: []projects.ProjectsData{{Title: "API<Server>", Excerpt: "CLI & Tools"}},
	}

	social := &links.LinksResponse{
		Data: []links.LinksData{{
			Name:        "X",
			Handle:      "@<gocanto>",
			URL:         "https://social.example/<gocanto>",
			Description: "Follow <me>",
		}},
	}

	recommendationsResponse := &recommendations.RecommendationsResponse{
		Data: []recommendations.RecommendationsData{{
			Relation: "Colleague <and> friend",
			Text:     "Great<br/>lead",
			Person: recommendations.RecommendationsPersonData{
				FullName:    "Jane <Doe>",
				Company:     "Tech <Corp>",
				Designation: "CTO",
			},
		}},
	}

	experienceResponse := &experience.ExperienceResponse{
		Data: []experience.ExperienceData{{
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

	educationResponse := &education.EducationResponse{
		Data: []education.EducationData{{
			School:         "Uni<versity>",
			Degree:         "BSc",
			Field:          "Computer <Science>",
			Description:    "Studied <algorithms>",
			GraduatedAt:    "2012",
			IssuingCountry: "Vene<zuela>",
		}},
	}

	categoriesList := []string{"Go<Lang>", "CLI"}

	publishedAt := time.Date(2024, time.January, 15, 0, 0, 0, 0, time.UTC)
	post := &posts.PostResponse{
		Title:   "Building <APIs>",
		Excerpt: "Learn <fast>\nwith examples",
		Content: "Intro paragraph with <tags>\nmore info.\n\nSecond paragraph & details.",
		Author: posts.UserResponse{
			DisplayName: "Gus <C>",
			Username:    "gocanto<script>",
			FirstName:   "Gus<",
			LastName:    "C>",
		},
		Categories:  []categoriespkg.CategoryResponse{{Name: "Go<Lang>"}},
		Tags:        []categoriespkg.TagResponse{{Name: "SEO<Tag>"}},
		PublishedAt: &publishedAt,
	}

	renderedProfile := string(sections.Profile(profileResponse))
	if strings.Contains(renderedProfile, "<Gus>") {
		t.Fatalf("profile output was not escaped: %q", renderedProfile)
	}

	renderedCategories := string(sections.Categories(categoriesList))
	if strings.Count(renderedCategories, "<li>") != 2 {
		t.Fatalf("expected two categories list items: %q", renderedCategories)
	}

	renderedSkills := string(sections.Skills(profileResponse))
	if !strings.Contains(renderedSkills, "Go&lt;lang&gt;") {
		t.Fatalf("skills should escape html: %q", renderedSkills)
	}

	renderedTalks := string(sections.Talks(talksResponse))
	if strings.Contains(renderedTalks, "Intro<Go>") {
		t.Fatalf("talks output was not escaped: %q", renderedTalks)
	}

	renderedProjects := string(sections.Projects(projectsResponse))
	if !strings.Contains(renderedProjects, "API&lt;Server&gt;") {
		t.Fatalf("projects output missing escaped title: %q", renderedProjects)
	}

	renderedSocial := string(sections.Social(social))
	if strings.Contains(renderedSocial, "<gocanto>") {
		t.Fatalf("social section should escape handles: %q", renderedSocial)
	}

	renderedRecommendations := string(sections.Recommendations(recommendationsResponse))
	if !strings.Contains(renderedRecommendations, "Great<br/>lead") {
		t.Fatalf("recommendations should allow br tags: %q", renderedRecommendations)
	}

	renderedExperience := string(sections.Experience(experienceResponse))
	if strings.Contains(renderedExperience, "Perx <Tech>") {
		t.Fatalf("experience should escape HTML: %q", renderedExperience)
	}

	renderedEducation := string(sections.Education(educationResponse))
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
	if strings.Contains(renderedPost, "By ") {
		t.Fatalf("post meta should not include author byline: %q", renderedPost)
	}
	if !strings.Contains(renderedPost, "Learn &lt;fast&gt;<br/>") {
		t.Fatalf("post excerpt should escape html and keep breaks: %q", renderedPost)
	}
	if !strings.Contains(renderedPost, "Second paragraph &amp; details.") {
		t.Fatalf("post content should escape html entities: %q", renderedPost)
	}

	for _, html := range []template.HTML{
		sections.Profile(profileResponse),
		sections.Categories(categoriesList),
		sections.Skills(profileResponse),
		sections.Talks(talksResponse),
		sections.Projects(projectsResponse),
		sections.Social(social),
		sections.Recommendations(recommendationsResponse),
		sections.Experience(experienceResponse),
		sections.Education(educationResponse),
		sections.Post(post),
	} {
		if !strings.HasPrefix(string(html), "<h1>") {
			t.Fatalf("section should start with heading: %q", html)
		}
	}
}

func TestSectionsGuardNilInputs(t *testing.T) {
	sections := seo.NewSections()

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
