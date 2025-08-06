package handler

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/oullin/database/repository"
)

func TestPostsHandlerIndex_ParseError(t *testing.T) {
	h := PostsHandler{Posts: &repository.Posts{}}
	badReq := httptest.NewRequest("POST", "/posts", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	if h.Index(rec, badReq) == nil {
		t.Fatalf("expected parse error")
	}
}

func TestPostsHandlerShow_MissingSlug(t *testing.T) {
	h := PostsHandler{Posts: &repository.Posts{}}
	req := httptest.NewRequest("GET", "/posts/", nil)
	rec := httptest.NewRecorder()
	if h.Show(rec, req) == nil {
		t.Fatalf("expected bad request")
	}
}
