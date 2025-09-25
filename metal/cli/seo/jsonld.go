package seo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"time"
)

type JsonID struct {
	SiteURL     string
	OrgName     string
	LogoURL     string
	Lang        string
	FoundedYear string
	SameAs      []string
	Now         func() time.Time

	// Repos and API
	WebRepoURL string
	APIRepoURL string
	APIName    string
	WebName    string
}

func NewJsonID(tmpl Page) *JsonID {
	return &JsonID{
		Lang:        tmpl.Lang,
		SiteURL:     tmpl.SiteURL,
		LogoURL:     tmpl.LogoURL,
		OrgName:     tmpl.SiteName,
		WebName:     tmpl.SiteName,
		APIName:     tmpl.SiteName,
		SameAs:      tmpl.SameAsURL,
		APIRepoURL:  tmpl.APIRepoURL,
		WebRepoURL:  tmpl.WebRepoURL,
		FoundedYear: fmt.Sprintf("%d", FoundedYear),
		Now:         func() time.Time { return time.Now().UTC() },
	}
}

func (j *JsonID) Render() template.JS {
	now := j.Now().Format(time.RFC3339)
	siteID := j.SiteURL + "#org"

	graph := []any{

		map[string]any{
			"@id":         siteID,
			"sameAs":      j.SameAs,
			"brand":       j.OrgName,
			"image":       j.LogoURL,
			"name":        j.OrgName,
			"legalName":   j.OrgName,
			"url":         j.SiteURL,
			"foundedYear": j.FoundedYear,
			"@type":       "Organization",
			"logo":        map[string]any{"@type": "ImageObject", "url": j.LogoURL},
		},

		map[string]any{
			"dateModified": now,
			"inLanguage":   j.Lang,
			"@type":        "WebSite",
			"url":          j.SiteURL,
			"name":         j.OrgName,
			"image":        j.LogoURL,
			"@id":          j.SiteURL + "#website",
			"publisher":    map[string]any{"@id": siteID},
		},

		map[string]any{
			"operatingSystem":     "All",
			"url":                 j.SiteURL,
			"name":                j.OrgName,
			"@type":               "WebApplication",
			"@id":                 j.SiteURL + "#app",
			"applicationCategory": "DeveloperApplication",
			"browserRequirements": "Requires a modern browser",
			"publisher":           map[string]any{"@id": siteID},
		},

		map[string]any{
			"@type":               "SoftwareSourceCode",
			"@id":                 j.WebRepoURL + "#code",
			"name":                j.WebName,
			"url":                 j.WebRepoURL,
			"codeRepository":      j.WebRepoURL,
			"programmingLanguage": []string{"TypeScript", "JavaScript", "Vue"},
			"issueTracker":        j.WebRepoURL + "/issues",
			"license":             j.WebRepoURL + "/blob/main/LICENSE",
			"publisher":           map[string]any{"@id": siteID},
			"dateModified":        now,
		},

		map[string]any{
			"dateModified":  now,
			"inLanguage":    j.Lang,
			"@type":         "WebAPI",
			"name":          j.APIName,
			"documentation": j.APIRepoURL,
			"@id":           j.SiteURL + "#api",
			"provider":      map[string]any{"@id": siteID},
			"softwareHelp": []any{
				map[string]any{
					"@type":                "CreativeWork",
					"learningResourceType": "Documentation",
					"url":                  j.APIRepoURL + "#readme",
				},
			},
			"workExample": []any{
				map[string]any{
					"dateModified":        now,
					"name":                j.APIName,
					"url":                 j.APIRepoURL,
					"codeRepository":      j.APIRepoURL,
					"@type":               "SoftwareSourceCode",
					"@id":                 j.APIRepoURL + "#code",
					"issueTracker":        j.APIRepoURL + "/issues",
					"programmingLanguage": []string{"Go", "Makefile"},
					"publisher":           map[string]any{"@id": siteID},
					"license":             j.APIRepoURL + "/blob/main/LICENSE",
				},
			},
		},
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
