package seo

import (
	"path"
	"strings"
)

func normalizeRelativeURL(rel string) string {
	rel = strings.ReplaceAll(rel, "\\", "/")

	cleaned := path.Clean(rel)

	if cleaned == "." || cleaned == "/" {
		return ""
	}

	parts := strings.Split(cleaned, "/")

	var b strings.Builder

	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			continue
		}

		if b.Len() > 0 {
			b.WriteByte('/')
		}

		b.WriteString(part)
	}

	return b.String()
}
