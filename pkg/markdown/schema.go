package markdown

import (
	"fmt"

	"github.com/oullin/pkg/portal"

	"time"
)

type FrontMatter struct {
	Title       string   `yaml:"title"`
	Excerpt     string   `yaml:"excerpt"`
	Slug        string   `yaml:"slug"`
	Author      string   `yaml:"author"`
	Categories  string   `yaml:"categories"`
	PublishedAt string   `yaml:"published_at"`
	Tags        []string `yaml:"tags"`
}

type Post struct {
	FrontMatter
	ImageURL string
	ImageAlt string
	Content  string
}

type Parser struct {
	Url string
}

func (f FrontMatter) GetPublishedAt() (*time.Time, error) {
	stringable := portal.NewStringable(f.PublishedAt)
	publishedAt, err := stringable.ToDatetime()

	if err != nil {
		return nil, fmt.Errorf("error parsing published_at: %v", err)
	}

	return publishedAt, nil
}
