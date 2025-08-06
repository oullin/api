package payload

import "testing"

func TestGetPostsFiltersFrom(t *testing.T) {
	req := IndexRequestBody{
		Title:    "t",
		Author:   "a",
		Category: "c",
		Tag:      "g",
		Text:     "x",
	}
	f := GetPostsFiltersFrom(req)
	if f.Title != "t" || f.Author != "a" || f.Category != "c" || f.Tag != "g" || f.Text != "x" {
		t.Fatalf("unexpected filters: %+v", f)
	}
}
