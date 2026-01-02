package payload_test

import (
	"testing"

	"github.com/oullin/handler/payload"
)

func TestGetPostsFiltersFrom(t *testing.T) {
	req := payload.IndexRequestBody{
		Title:    "t",
		Author:   "a",
		Category: "c",
		Tag:      "g",
		Text:     "x",
	}

	f := payload.GetPostsFiltersFrom(req)

	if f.Title != "t" || f.Author != "a" || f.Category != "c" || f.Tag != "g" || f.Text != "x" {
		t.Fatalf("unexpected filters: %+v", f)
	}
}
