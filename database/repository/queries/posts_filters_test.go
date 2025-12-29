package queries_test

import (
	"testing"

	"github.com/oullin/database/repository/queries"
)

func TestPostFiltersSanitise(t *testing.T) {
	f := queries.PostFilters{
		Text:     "  Hello  ",
		Title:    "  MyTitle  ",
		Author:   "  ME  ",
		Category: "  Tech  ",
		Tag:      "Tag  ",
	}

	if f.GetText() != "hello" {
		t.Fatalf("expected GetText to return 'hello', got %s", f.GetText())
	}

	if f.GetTitle() != "mytitle" {
		t.Fatalf("expected GetTitle to return 'mytitle', got %s", f.GetTitle())
	}

	if f.GetAuthor() != "me" {
		t.Fatalf("expected GetAuthor to return 'me', got %s", f.GetAuthor())
	}

	if f.GetCategory() != "tech" {
		t.Fatalf("expected GetCategory to return 'tech', got %s", f.GetCategory())
	}

	if f.GetTag() != "tag" {
		t.Fatalf("expected GetTag to return 'tag', got %s", f.GetTag())
	}
}
