package posts

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	clitest "github.com/oullin/cli/clitest"
	"github.com/oullin/database"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/markdown"
)

func captureOutput(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = old
	return string(out)
}

func setupPostsHandler(t *testing.T) (*Handler, *database.Connection) {
	conn := clitest.MakeTestConnection(t, &database.User{}, &database.Post{}, &database.Category{}, &database.PostCategory{}, &database.Tag{}, &database.PostTag{})
	user := database.User{UUID: uuid.NewString(), Username: "jdoe", FirstName: "John", LastName: "Doe", Email: "jdoe@example.com", PasswordHash: "x"}
	if err := conn.Sql().Create(&user).Error; err != nil {
		t.Fatalf("user create: %v", err)
	}
	conn.Sql().Create(&database.Category{UUID: uuid.NewString(), Name: "Tech", Slug: "tech"})
	conn.Sql().Create(&database.Tag{UUID: uuid.NewString(), Name: "Go", Slug: "go"})
	input := &Input{Url: "http://example"}
	h := MakeHandler(input, pkg.MakeDefaultClient(nil), conn)
	return &h, conn
}

func TestHandlePost(t *testing.T) {
	h, conn := setupPostsHandler(t)
	post := &markdown.Post{
		FrontMatter: markdown.FrontMatter{
			Title:       "Hello",
			Excerpt:     "ex",
			Slug:        "hello",
			Author:      "jdoe",
			Categories:  "tech",
			PublishedAt: time.Now().Format("2006-01-02"),
			Tags:        []string{"go"},
		},
		Content: "world",
	}
	if err := h.HandlePost(post); err != nil {
		t.Fatalf("handle: %v", err)
	}
	var p database.Post
	if err := conn.Sql().Preload("Categories").First(&p, "slug = ?", "hello").Error; err != nil {
		t.Fatalf("post not created: %v", err)
	}
	if len(p.Categories) != 1 {
		t.Fatalf("expected 1 category")
	}
	_ = captureOutput(func() { h.RenderArticle(post) })
}

func TestHandlePostMissingAuthor(t *testing.T) {
	h, _ := setupPostsHandler(t)
	post := &markdown.Post{FrontMatter: markdown.FrontMatter{Author: "none"}}
	if err := h.HandlePost(post); err == nil {
		t.Fatalf("expected error")
	}
}

func TestHandlePostEmptyCategories(t *testing.T) {
	h, _ := setupPostsHandler(t)
	post := &markdown.Post{FrontMatter: markdown.FrontMatter{Author: "jdoe"}}
	if err := h.HandlePost(post); err == nil {
		t.Fatalf("expected error")
	}
}

func TestHandlePostInvalidDate(t *testing.T) {
	h, _ := setupPostsHandler(t)
	post := &markdown.Post{FrontMatter: markdown.FrontMatter{Author: "jdoe", PublishedAt: "bad"}}
	if err := h.HandlePost(post); err == nil {
		t.Fatalf("expected error")
	}
}

func TestHandlePostDuplicateSlug(t *testing.T) {
	h, _ := setupPostsHandler(t)
	post := &markdown.Post{FrontMatter: markdown.FrontMatter{Author: "jdoe", Slug: "dup", Categories: "tech", PublishedAt: time.Now().Format("2006-01-02")}}
	if err := h.HandlePost(post); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if err := h.HandlePost(post); err == nil {
		t.Fatalf("expected duplicate error")
	}
}

func TestNotParsed(t *testing.T) {
	h, conn := setupPostsHandler(t)
	md := "---\nauthor: jdoe\nslug: parsed\ncategories: tech\npublished_at: 2024-01-01\ntags:\n - go\n---\ncontent"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(md)) }))
	defer srv.Close()
	h.Input.Url = srv.URL
	ok, err := h.NotParsed()
	if err != nil || !ok {
		t.Fatalf("not parsed: %v", err)
	}
	var p database.Post
	if err := conn.Sql().First(&p, "slug = ?", "parsed").Error; err != nil {
		t.Fatalf("post not saved")
	}
}

func TestNotParsedError(t *testing.T) {
	h, _ := setupPostsHandler(t)
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	h.Input.Url = srv.URL
	if ok, err := h.NotParsed(); err == nil || ok {
		t.Fatalf("expected error")
	}
}
