package projects

import (
	"sort"
	"strings"
	"time"

	"github.com/oullin/handler/payload"
)

func SortBySortAsc(projects []payload.ProjectsData) {
	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].Sort != projects[j].Sort {
			return projects[i].Sort < projects[j].Sort
		}

		left, leftOK := publishedAtDate(projects[i])
		right, rightOK := publishedAtDate(projects[j])

		switch {
		case leftOK && rightOK:
			return left.After(right)
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return false
		}
	})
}

func publishedAtDate(project payload.ProjectsData) (time.Time, bool) {
	if parsed, ok := ParsePublishedAt(project.PublishedAt); ok {
		return parsed, true
	}

	return time.Time{}, false
}

func ParsePublishedAt(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return time.Time{}, false
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed, true
		}
	}

	return time.Time{}, false
}
