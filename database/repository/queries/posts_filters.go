package queries

import (
	"github.com/oullin/pkg"
	"strings"
)

type PostFilters struct {
	Text        string
	Title       string // Will perform a case-insensitive partial match
	Author      string
	Category    string
	Tag         string
	IsPublished *bool // Pointer to bool to allow three states: true, false, and not-set (nil)
}

func (f PostFilters) GetText() string {
	return f.sanitiseString(f.Text)
}

func (f PostFilters) GetTitle() string {
	return f.sanitiseString(f.Title)
}

func (f PostFilters) GetAuthor() string {
	return f.sanitiseString(f.Author)
}

func (f PostFilters) GetCategory() string {
	return f.sanitiseString(f.Category)
}

func (f PostFilters) GetTag() string {
	return f.sanitiseString(f.Tag)
}

func (f PostFilters) sanitiseString(seed string) string {
	str := pkg.MakeStringable(seed)

	return strings.TrimSpace(str.ToLower())
}
