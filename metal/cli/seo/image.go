package seo

import (
	"path"
	"strings"
)

func normalizeRelativeURL(rel string) string {
	cleaned := path.Clean(rel)
	cleaned = strings.TrimPrefix(cleaned, "./")

	for strings.HasPrefix(cleaned, "../") {
		cleaned = strings.TrimPrefix(cleaned, "../")
	}

	cleaned = strings.TrimPrefix(cleaned, "/")

	return cleaned
}
