package seo

import (
	"path"
	"strings"
)

func normalizeRelativeURL(rel string) string {
	cleaned := path.Clean(rel)

	if cleaned == "." {
		return ""
	}

	parts := strings.Split(cleaned, "/")
	normalized := make([]string, 0, len(parts))

	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			continue
		}

		normalized = append(normalized, part)
	}

	return strings.Join(normalized, "/")
}
