package seo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"time"
)

type JsonID struct {
	SiteURL         string
	OrgName         string
	LogoURL         string
	Lang            string
	FoundedYear     string
	SameAs          []string
	SiteDescription string
	PageName        string
	PageType        string
	PageURL         string
	PageDescription string
	Founder         *JsonPerson
	Now             func() time.Time

	// Repos and API
	WebRepoURL string
	APIRepoURL string
	APIName    string
	WebName    string
}

type JsonPerson struct {
	Name        string
	JobTitle    string
	URL         string
	Description string
}

func NewJsonID(tmpl Page, web *Web) *JsonID {
	if web == nil {
		web = NewWeb()
	}

	return &JsonID{
		Lang:            tmpl.Lang,
		SiteURL:         tmpl.SiteURL,
		LogoURL:         tmpl.LogoURL,
		OrgName:         tmpl.SiteName,
		WebName:         tmpl.SiteName,
		APIName:         tmpl.SiteName,
		SameAs:          tmpl.SameAsURL,
		APIRepoURL:      tmpl.APIRepoURL,
		WebRepoURL:      tmpl.WebRepoURL,
		FoundedYear:     fmt.Sprintf("%d", web.FoundedYear),
		SiteDescription: web.Description,
		Now:             func() time.Time { return time.Now().UTC() },
	}
}

func (j *JsonID) WithPage(name, pageType, url, description string) *JsonID {
	j.PageName = name
	j.PageType = pageType
	j.PageURL = url
	j.PageDescription = description

	return j
}

func (j *JsonID) WithFounder(person JsonPerson) *JsonID {
	j.Founder = &person

	return j
}

func (j *JsonID) Render() template.JS {
	siteID := j.SiteURL + "#org"
	websiteID := j.SiteURL + "#website"

	graph := []any{
		map[string]any{
			"@id":         siteID,
			"sameAs":      j.SameAs,
			"brand":       j.OrgName,
			"image":       j.LogoURL,
			"name":        j.OrgName,
			"legalName":   j.OrgName,
			"url":         j.SiteURL,
			"description": j.SiteDescription,
			"foundedYear": j.FoundedYear,
			"@type":       "Organization",
			"logo":        map[string]any{"@type": "ImageObject", "url": j.LogoURL},
		},

		map[string]any{
			"inLanguage":  j.Lang,
			"@type":       "WebSite",
			"url":         j.SiteURL,
			"name":        j.OrgName,
			"image":       j.LogoURL,
			"description": j.SiteDescription,
			"@id":         websiteID,
			"publisher":   map[string]any{"@id": siteID},
		},
	}

	if j.PageType != "" && j.PageURL != "" {
		graph = append(graph, map[string]any{
			"@context":    "https://schema.org",
			"@type":       j.PageType,
			"name":        j.PageName,
			"url":         j.PageURL,
			"description": j.PageDescription,
			"isPartOf":    map[string]any{"@id": websiteID},
		})
	}

	if j.Founder != nil {
		founder := map[string]any{
			"@context": "https://schema.org",
			"@type":    "Person",
			"name":     j.Founder.Name,
			"jobTitle": j.Founder.JobTitle,
			"url":      j.Founder.URL,
		}

		if j.Founder.Description != "" {
			founder["description"] = j.Founder.Description
		}

		graph = append(graph, founder)
	}

	root := map[string]any{
		"@graph":   graph,
		"@context": "https://schema.org",
	}

	// Encode without Template escaping and compact.
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(root); err != nil {
		return `{}`
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, buf.Bytes()); err != nil {
		// Fallback to un-compacted if compaction fails
		return template.JS(buf.String())
	}

	return template.JS(compact.String())
}
