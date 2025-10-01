package seo

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/portal"
)

type Sections struct{}

func NewSections() Sections {
	return Sections{}
}

func (s *Sections) Categories(categories []string) template.HTML {
	var items []string

	for _, item := range categories {
		items = append(items, "<li>"+template.HTMLEscapeString(item)+"</li>")
	}

	return template.HTML("<h1>Categories</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) Profile(profile *payload.ProfileResponse) template.HTML {
	if profile == nil {
		return template.HTML("")
	}

	return "<h1>Profile</h1>" +
		template.HTML("<p>"+
			template.HTMLEscapeString(profile.Data.Name)+", "+
			template.HTMLEscapeString(profile.Data.Profession)+
			"</p>",
		)
}

func (s *Sections) Skills(profile *payload.ProfileResponse) template.HTML {
	if profile == nil {
		return template.HTML("")
	}

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
	if talks == nil {
		return template.HTML("")
	}

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
		title := template.HTMLEscapeString(item.Title)
		lang := template.HTMLEscapeString(item.Language)
		excerpt := template.HTMLEscapeString(item.Excerpt)
		href := portal.SanitizeURL(strings.TrimSpace(item.URL))

		project := fmt.Sprintf("<strong>%s</strong>", title)

		if href != "" {
			project = fmt.Sprintf("<a href=\"%s\">%s</a>", href, project)
		}

		details := []string{}
		if excerpt != "" {
			details = append(details, excerpt)
		}

		if lang != "" {
			details = append(details, "Language: "+lang)
		}

		items = append(items, "<li>"+
			project+
			s.FormatDetails(details)+
			"</li>",
		)
	}

	return template.HTML("<h1>Projects</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) Post(post *payload.PostResponse) template.HTML {
	if post == nil {
		return template.HTML("")
	}

	title := template.HTMLEscapeString(post.Title)

	authorName := strings.TrimSpace(post.Author.DisplayName)
	if authorName == "" {
		fullName := strings.TrimSpace(strings.Join(portal.FilterNonEmpty(
			[]string{post.Author.FirstName, post.Author.LastName}), " "),
		)

		if fullName != "" {
			authorName = fullName
		} else {
			authorName = strings.TrimSpace(post.Author.Username)
		}
	}

	authorName = template.HTMLEscapeString(authorName)

	var metaParts []string
	if authorName != "" {
		metaParts = append(metaParts, "By "+authorName)
	}

	if post.PublishedAt != nil {
		published := post.PublishedAt.UTC().Format("02 Jan 2006")
		metaParts = append(metaParts, "Published "+template.HTMLEscapeString(published))
	}

	if len(post.Categories) > 0 {
		var names []string

		for _, category := range post.Categories {
			name := strings.TrimSpace(category.Name)
			if name != "" {
				names = append(names, template.HTMLEscapeString(name))
			}
		}

		if len(names) > 0 {
			metaParts = append(metaParts, "Categories: "+strings.Join(names, ", "))
		}
	}

	if len(post.Tags) > 0 {
		var names []string

		for _, tag := range post.Tags {
			name := strings.TrimSpace(tag.Name)
			if name != "" {
				names = append(names, template.HTMLEscapeString(name))
			}
		}

		if len(names) > 0 {
			metaParts = append(metaParts, "Tags: "+strings.Join(names, ", "))
		}
	}

	metaHTML := ""
	if len(metaParts) > 0 {
		metaHTML = "<p><small>" + strings.Join(metaParts, " | ") + "</small></p>"
	}

	excerpt := strings.TrimSpace(post.Excerpt)
	excerptHTML := ""

	if excerpt != "" {
		escaped := template.HTMLEscapeString(strings.ReplaceAll(excerpt, "\r\n", "\n"))
		escaped = strings.ReplaceAll(escaped, "\n", "<br/>")
		excerptHTML = "<p>" + portal.AllowLineBreaks(escaped) + "</p>"
	}

	contentHTML := s.FormatPostContent(post.Content)

	return template.HTML("<h1>" + title + "</h1>" + metaHTML + excerptHTML + contentHTML)
}

func (s *Sections) Social(social *payload.SocialResponse) template.HTML {
	if social == nil {
		return template.HTML("<h1>Social</h1><p><ul></ul></p>")
	}

	var items []string

	for _, item := range social.Data {
		name := template.HTMLEscapeString(item.Name)
		handle := template.HTMLEscapeString(item.Handle)
		href := portal.SanitizeURL(strings.TrimSpace(item.URL))
		description := template.HTMLEscapeString(item.Description)

		linkText := name
		if handle != "" {
			linkText = fmt.Sprintf("%s (%s)", linkText, handle)
		}

		if href != "" {
			linkText = fmt.Sprintf("<a href=\"%s\">%s</a>", href, linkText)
		}

		items = append(items, "<li>"+
			linkText+
			s.FormatDetails([]string{description})+
			"</li>",
		)
	}

	return template.HTML("<h1>Social</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) Recommendations(recs *payload.RecommendationsResponse) template.HTML {
	if recs == nil {
		return template.HTML("<h1>Recommendations</h1><p><ul></ul></p>")
	}

	var items []string

	for _, item := range recs.Data {
		relation := template.HTMLEscapeString(item.Relation)
		company := template.HTMLEscapeString(item.Person.Company)
		fullName := template.HTMLEscapeString(item.Person.FullName)
		designation := template.HTMLEscapeString(item.Person.Designation)
		text := portal.AllowLineBreaks(template.HTMLEscapeString(item.Text))

		meta := []string{}
		if designation != "" {
			meta = append(meta, designation)
		}

		if company != "" {
			meta = append(meta, company)
		}

		heading := fullName
		if len(meta) > 0 {
			heading += " (" + strings.Join(meta, ", ") + ")"
		}

		details := []string{}
		if relation != "" {
			details = append(details, relation)
		}

		if text != "" {
			details = append(details, text)
		}

		items = append(items, "<li>"+
			"<strong>"+heading+"</strong>"+
			s.FormatDetails(details)+
			"</li>",
		)
	}

	return template.HTML("<h1>Recommendations</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) Experience(exp *payload.ExperienceResponse) template.HTML {
	if exp == nil {
		return template.HTML("<h1>Experience</h1><p><ul></ul></p>")
	}

	var items []string

	for _, item := range exp.Data {
		city := template.HTMLEscapeString(item.City)
		end := template.HTMLEscapeString(item.EndDate)
		skills := template.HTMLEscapeString(item.Skills)
		company := template.HTMLEscapeString(item.Company)
		country := template.HTMLEscapeString(item.Country)
		start := template.HTMLEscapeString(item.StartDate)
		summary := template.HTMLEscapeString(item.Summary)
		position := template.HTMLEscapeString(item.Position)
		locationType := template.HTMLEscapeString(item.LocationType)
		employmentType := template.HTMLEscapeString(item.EmploymentType)

		timeline := strings.Join(portal.FilterNonEmpty([]string{start, end}), " - ")
		location := strings.Join(portal.FilterNonEmpty([]string{city, country}), ", ")
		heading := strings.Join(portal.FilterNonEmpty([]string{position, company}), " at ")

		details := []string{}
		if timeline != "" {
			details = append(details, "Timeline: "+timeline)
		}

		if employmentType != "" || locationType != "" {
			details = append(details, strings.Join(portal.FilterNonEmpty([]string{employmentType, locationType}), " · "))
		}

		if location != "" {
			details = append(details, "Location: "+location)
		}

		if summary != "" {
			details = append(details, summary)
		}

		if skills != "" {
			details = append(details, "Skills: "+skills)
		}

		items = append(items, "<li>"+
			"<strong>"+heading+"</strong>"+
			s.FormatDetails(details)+
			"</li>",
		)
	}

	return template.HTML("<h1>Experience</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) Education(edu *payload.EducationResponse) template.HTML {
	if edu == nil {
		return template.HTML("<h1>Education</h1><p><ul></ul></p>")
	}

	var items []string

	for _, item := range edu.Data {
		field := template.HTMLEscapeString(item.Field)
		school := template.HTMLEscapeString(item.School)
		degree := template.HTMLEscapeString(item.Degree)
		graduated := template.HTMLEscapeString(item.GraduatedAt)
		description := template.HTMLEscapeString(item.Description)
		country := template.HTMLEscapeString(item.IssuingCountry)

		headingParts := portal.FilterNonEmpty([]string{degree, field})
		heading := strings.Join(headingParts, " in ")

		if heading == "" {
			heading = school
		} else if school != "" {
			heading += " at " + school
		}

		details := []string{}
		if graduated != "" {
			details = append(details, "Graduated: "+graduated)
		}

		if country != "" {
			details = append(details, "Country: "+country)
		}

		if description != "" {
			details = append(details, portal.AllowLineBreaks(description))
		}

		items = append(items, "<li>"+
			"<strong>"+heading+"</strong>"+
			s.FormatDetails(details)+
			"</li>",
		)
	}

	return template.HTML("<h1>Education</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) FormatPostContent(content string) string {
	trimmed := strings.TrimSpace(strings.ReplaceAll(content, "\r\n", "\n"))
	if trimmed == "" {
		return ""
	}

	rawParagraphs := strings.Split(trimmed, "\n\n")
	var rendered []string

	for _, paragraph := range rawParagraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		escaped := template.HTMLEscapeString(paragraph)
		escaped = strings.ReplaceAll(escaped, "\n", "<br/>")
		escaped = portal.AllowLineBreaks(escaped)
		rendered = append(rendered, "<p>"+escaped+"</p>")
	}

	if len(rendered) == 0 {
		return ""
	}

	return strings.Join(rendered, "")
}

func (s *Sections) FormatDetails(parts []string) string {
	filtered := portal.FilterNonEmpty(parts)

	if len(filtered) == 0 {
		return ""
	}

	return ": " + strings.Join(filtered, " | ")
}
