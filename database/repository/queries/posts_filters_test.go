package queries

import "testing"

func TestPostFiltersSanitise(t *testing.T) {
	f := PostFilters{
		Text:     "  Hello  ",
		Title:    "  MyTitle  ",
		Author:   "  ME  ",
		Category: "  Tech  ",
		Tag:      "Tag  ",
	}

	if f.GetText() != "hello" {
		t.Fatalf("got %s", f.GetText())
	}
	if f.GetTitle() != "mytitle" {
		t.Fatalf("got %s", f.GetTitle())
	}
	if f.GetAuthor() != "me" {
		t.Fatalf("got %s", f.GetAuthor())
	}
	if f.GetCategory() != "tech" {
		t.Fatalf("got %s", f.GetCategory())
	}
	if f.GetTag() != "tag" {
		t.Fatalf("got %s", f.GetTag())
	}
}
