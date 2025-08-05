package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/database/repository/queries"
	"github.com/oullin/handler/payload"
)

func TestPostsHandlerIndex(t *testing.T) {
	post := database.Post{UUID: "p1", Slug: "slug", Title: "title"}
	pag := pagination.Paginate{Page: 1, Limit: 10}
	pag.SetNumItems(1)
	list := pagination.MakePagination([]database.Post{post}, pag)
	repoErr := error(nil)
	h := MakePostsHandler(
		func(filters queries.PostFilters, p pagination.Paginate) (*pagination.Pagination[database.Post], error) {
			return list, repoErr
		},
		func(slug string) *database.Post { return &post },
	)

	body, _ := json.Marshal(payload.IndexRequestBody{Title: "title"})
	req := httptest.NewRequest("POST", "/posts", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	if err := h.Index(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	repoErr = errors.New("fail")
	rec2 := httptest.NewRecorder()
	if h.Index(rec2, req) == nil {
		t.Fatalf("expected error")
	}

	badReq := httptest.NewRequest("POST", "/posts", bytes.NewReader([]byte("{")))
	rec3 := httptest.NewRecorder()
	if h.Index(rec3, badReq) == nil {
		t.Fatalf("expected parse error")
	}
}

func TestPostsHandlerShow(t *testing.T) {
	post := database.Post{UUID: "p1", Slug: "slug", Title: "title"}
	item := &post
	h := MakePostsHandler(
		func(filters queries.PostFilters, p pagination.Paginate) (*pagination.Pagination[database.Post], error) {
			return nil, nil
		},
		func(slug string) *database.Post { return item },
	)

	req := httptest.NewRequest("GET", "/posts/slug", nil)
	req.SetPathValue("slug", "slug")
	rec := httptest.NewRecorder()
	if err := h.Show(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	req2 := httptest.NewRequest("GET", "/posts/", nil)
	rec2 := httptest.NewRecorder()
	if h.Show(rec2, req2) == nil {
		t.Fatalf("expected bad request")
	}

	item = nil
	req3 := httptest.NewRequest("GET", "/posts/slug", nil)
	req3.SetPathValue("slug", "slug")
	rec3 := httptest.NewRecorder()
	if h.Show(rec3, req3) == nil {
		t.Fatalf("expected not found")
	}
}
