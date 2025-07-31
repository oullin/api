package payload

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oullin/database"
)

func TestGetPostsFiltersFrom(t *testing.T) {
	req := IndexRequestBody{Title: "t", Author: "a", Category: "c", Tag: "g", Text: "x"}
	f := GetPostsFiltersFrom(req)
	if f.Title != "t" || f.Author != "a" || f.Category != "c" || f.Tag != "g" || f.Text != "x" {
		t.Fatalf("unexpected filters: %+v", f)
	}
}

func TestGetSlugFrom(t *testing.T) {
	r := httptest.NewRequest("GET", "/posts/s", nil)
	r.SetPathValue("slug", "  SLUG ")
	if s := GetSlugFrom(r); s != "slug" {
		t.Fatalf("slug %s", s)
	}
}

func TestGetPostsResponse(t *testing.T) {
	now := time.Now()
	p := database.Post{
		UUID: "1", Slug: "slug", Title: "title", Excerpt: "ex", Content: "c",
		CoverImageURL: "url", PublishedAt: &now, CreatedAt: now, UpdatedAt: now,
		Categories: []database.Category{{UUID: "c1", Name: "cn", Slug: "cs", Description: "cd"}},
		Tags:       []database.Tag{{UUID: "t1", Name: "tn", Slug: "ts", Description: "td"}},
		Author:     database.User{UUID: "u1", FirstName: "fn", LastName: "ln", Username: "un", DisplayName: "dn", Bio: "b", PictureFileName: "pf", ProfilePictureURL: "pu", IsAdmin: true},
	}

	r := GetPostsResponse(p)
	if r.UUID != "1" || r.Author.UUID != "u1" || len(r.Categories) != 1 || len(r.Tags) != 1 {
		t.Fatalf("unexpected response: %+v", r)
	}
}
