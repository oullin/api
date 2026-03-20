package markdown_test

import (
	"github.com/oullin/internal/shared/markdown"

	"testing"
)

func TestParseWithHeaderImage(t *testing.T) {
	md := `---
slug: test
published_at: 2024-06-09
---
![alt](url)
content`

	post, err := markdown.Parse(md)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if post.ImageAlt != "alt" || post.ImageURL != "url" || post.Content != "content" {
		t.Fatalf("parse failed")
	}

	if post.Slug != "test" {
		t.Fatalf("front matter parse failed")
	}

	if _, err := post.GetPublishedAt(); err != nil {
		t.Fatalf("get date: %v", err)
	}
}

func TestParseWithoutHeaderImage(t *testing.T) {
	md := `---
slug: another
published_at: 2024-06-09
---
content`

	post, err := markdown.Parse(md)

	if err != nil {
		t.Fatalf("err: %v", err)
	}

	if post.ImageAlt != "" || post.ImageURL != "" || post.Content != "content" {
		t.Fatalf("parse failed")
	}
}

func TestParseErrors(t *testing.T) {
	if _, err := markdown.Parse("invalid"); err == nil {
		t.Fatalf("expected error")
	}

	if _, err := markdown.Parse(`---\nbad`); err == nil {
		t.Fatalf("expected bad yaml error")
	}

	md := `---
slug: a
published_at: "bad"
---
content`
	post, err := markdown.Parse(md)

	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if _, err := post.GetPublishedAt(); err == nil {
		t.Fatalf("expected date error")
	}
}
