package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler/payload"
	handlertests "github.com/oullin/handler/tests"
)

func TestPostsHandlerIndex_ParseError(t *testing.T) {
	h := PostsHandler{
		Posts: &repository.Posts{},
	}

	badReq := httptest.NewRequest("POST", "/posts", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()

	if h.Index(rec, badReq) == nil {
		t.Fatalf("expected parse error")
	}
}

func TestPostsHandlerShow_MissingSlug(t *testing.T) {
	h := PostsHandler{
		Posts: &repository.Posts{},
	}

	req := httptest.NewRequest("GET", "/posts/", nil)
	rec := httptest.NewRecorder()

	if h.Show(rec, req) == nil {
		t.Fatalf("expected bad request")
	}
}

func TestPostsHandlerIndex_Success(t *testing.T) {
	conn, author := handlertests.NewTestDB(t)
	published := time.Now()
	post := database.Post{
		UUID:        uuid.NewString(),
		AuthorID:    author.ID,
		Slug:        "hello",
		Title:       "Hello",
		Excerpt:     "Ex",
		Content:     "Body",
		PublishedAt: &published,
	}

	if err := conn.Sql().Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	h := NewPostsHandler(&repository.Posts{
		DB: conn,
	})

	req := httptest.NewRequest("POST", "/posts", bytes.NewReader([]byte("{}")))
	rec := httptest.NewRecorder()

	if err := h.Index(rec, req); err != nil {
		t.Fatalf("index err: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var resp pagination.Pagination[payload.PostResponse]

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Data) != 1 || resp.Data[0].Slug != "hello" {
		t.Fatalf("unexpected data: %+v", resp.Data)
	}
}

func TestPostsHandlerShow_Success(t *testing.T) {
	conn, author := handlertests.NewTestDB(t)
	published := time.Now()
	tag := database.Tag{
		UUID: uuid.NewString(),
		Name: "Go",
		Slug: "go",
	}

	if err := conn.Sql().Create(&tag).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}

	post := database.Post{
		UUID:        uuid.NewString(),
		AuthorID:    author.ID,
		Slug:        "hello",
		Title:       "Hello",
		Excerpt:     "Ex",
		Content:     "Body",
		PublishedAt: &published,
	}

	if err := conn.Sql().Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	if err := conn.Sql().Create(&database.PostTag{PostID: post.ID, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("create post tag: %v", err)
	}

	h := NewPostsHandler(&repository.Posts{
		DB: conn,
	})

	req := httptest.NewRequest("GET", "/posts/hello", nil)
	req.SetPathValue("slug", "hello")
	rec := httptest.NewRecorder()

	if err := h.Show(rec, req); err != nil {
		t.Fatalf("show err: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var resp payload.PostResponse

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Slug != "hello" {
		t.Fatalf("unexpected slug: %s", resp.Slug)
	}

	if len(resp.Tags) != 1 || resp.Tags[0].Slug != "go" || resp.Tags[0].Name != "Go" {
		t.Fatalf("unexpected tags: %+v", resp.Tags)
	}
}
