package seo

import (
	"fmt"
	"html/template"
	"strings"

	"github.com/oullin/handler/payload"
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
		href := sanitizeURL(item.URL)
		title := template.HTMLEscapeString(item.Title)
		excerpt := template.HTMLEscapeString(item.Excerpt)
		lang := template.HTMLEscapeString(item.Language)

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
			formatDetails(details)+
			"</li>",
		)
	}

	return template.HTML("<h1>Projects</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func (s *Sections) Social(social *payload.SocialResponse) template.HTML {
	if social == nil {
		return template.HTML("<h1>Social</h1><p><ul></ul></p>")
	}

	var items []string

	for _, item := range social.Data {
		href := sanitizeURL(item.URL)
		name := template.HTMLEscapeString(item.Name)
		handle := template.HTMLEscapeString(item.Handle)
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
			formatDetails([]string{description})+
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
		fullName := template.HTMLEscapeString(item.Person.FullName)
		designation := template.HTMLEscapeString(item.Person.Designation)
		company := template.HTMLEscapeString(item.Person.Company)
		relation := template.HTMLEscapeString(item.Relation)
		text := allowLineBreaks(template.HTMLEscapeString(item.Text))

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
			formatDetails(details)+
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
		company := template.HTMLEscapeString(item.Company)
		position := template.HTMLEscapeString(item.Position)
		employmentType := template.HTMLEscapeString(item.EmploymentType)
		locationType := template.HTMLEscapeString(item.LocationType)
		city := template.HTMLEscapeString(item.City)
		country := template.HTMLEscapeString(item.Country)
		start := template.HTMLEscapeString(item.StartDate)
		end := template.HTMLEscapeString(item.EndDate)
		summary := template.HTMLEscapeString(item.Summary)
		skills := template.HTMLEscapeString(item.Skills)

		timeline := strings.Join(filterNonEmpty([]string{start, end}), " - ")
		location := strings.Join(filterNonEmpty([]string{city, country}), ", ")
		heading := strings.Join(filterNonEmpty([]string{position, company}), " at ")

		details := []string{}
		if timeline != "" {
			details = append(details, "Timeline: "+timeline)
		}
		if employmentType != "" || locationType != "" {
			details = append(details, strings.Join(filterNonEmpty([]string{employmentType, locationType}), " · "))
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
			formatDetails(details)+
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
		school := template.HTMLEscapeString(item.School)
		degree := template.HTMLEscapeString(item.Degree)
		field := template.HTMLEscapeString(item.Field)
		description := template.HTMLEscapeString(item.Description)
		graduated := template.HTMLEscapeString(item.GraduatedAt)
		country := template.HTMLEscapeString(item.IssuingCountry)

		headingParts := filterNonEmpty([]string{degree, field})
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
			details = append(details, allowLineBreaks(description))
		}

		items = append(items, "<li>"+
			"<strong>"+heading+"</strong>"+
			formatDetails(details)+
			"</li>",
		)
	}

	return template.HTML("<h1>Education</h1>" +
		"<p><ul>" +
		strings.Join(items, "") +
		"</ul></p>",
	)
}

func formatDetails(parts []string) string {
	filtered := filterNonEmpty(parts)

	if len(filtered) == 0 {
		return ""
	}

	return ": " + strings.Join(filtered, " | ")
}

func allowLineBreaks(text string) string {
	replacer := strings.NewReplacer(
		"&lt;br/&gt;", "<br/>",
		"&lt;br /&gt;", "<br/>",
		"&lt;br&gt;", "<br/>",
	)

	return replacer.Replace(text)
}

func filterNonEmpty(values []string) []string {
	var out []string
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, strings.TrimSpace(v))
		}
	}

	return out
}

// sanitizeURL only allows http(s) schemes and returns an escaped value or empty string.
func sanitizeURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	lower := strings.ToLower(u)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return template.HTMLEscapeString(u)
	}
	return ""
}
