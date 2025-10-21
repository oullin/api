package handler

import (
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

func TestCategoriesHandlerIndex_Success(t *testing.T) {
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

	catB := database.Category{
		UUID:        uuid.NewString(),
		Name:        "Beta",
		Slug:        "beta",
		Description: "desc",
		Sort:        10,
	}

	if err := conn.Sql().Create(&catB).Error; err != nil {
		t.Fatalf("create category beta: %v", err)
	}

	catA := database.Category{
		UUID:        uuid.NewString(),
		Name:        "Alpha",
		Slug:        "alpha",
		Description: "desc",
		Sort:        10,
	}

	if err := conn.Sql().Create(&catA).Error; err != nil {
		t.Fatalf("create category alpha: %v", err)
	}

	linkBeta := database.PostCategory{
		PostID:     post.ID,
		CategoryID: catB.ID,
	}

	if err := conn.Sql().Create(&linkBeta).Error; err != nil {
		t.Fatalf("create beta link: %v", err)
	}

	linkAlpha := database.PostCategory{
		PostID:     post.ID,
		CategoryID: catA.ID,
	}

	if err := conn.Sql().Create(&linkAlpha).Error; err != nil {
		t.Fatalf("create alpha link: %v", err)
	}

	h := NewCategoriesHandler(&repository.Categories{
		DB: conn,
	})

	req := httptest.NewRequest("GET", "/categories", nil)
	rec := httptest.NewRecorder()

	if err := h.Index(rec, req); err != nil {
		t.Fatalf("index err: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var resp pagination.Pagination[payload.CategoryResponse]

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("unexpected data length: %+v", resp.Data)
	}

	if resp.Data[0].Name != "Alpha" || resp.Data[0].Slug != "alpha" || resp.Data[0].Sort != 10 {
		t.Fatalf("unexpected first category: %+v", resp.Data[0])
	}

	if resp.Data[1].Name != "Beta" || resp.Data[1].Slug != "beta" || resp.Data[1].Sort != 10 {
		t.Fatalf("unexpected second category: %+v", resp.Data[1])
	}
}

func TestCategoriesHandlerIndex_SortOrdering(t *testing.T) {
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

	catLow := database.Category{
		UUID:        uuid.NewString(),
		Name:        "Zebra",
		Slug:        "zebra",
		Description: "desc",
		Sort:        5,
	}

	if err := conn.Sql().Create(&catLow).Error; err != nil {
		t.Fatalf("create low sort category: %v", err)
	}

	catHigh := database.Category{
		UUID:        uuid.NewString(),
		Name:        "Apple",
		Slug:        "apple",
		Description: "desc",
		Sort:        20,
	}

	if err := conn.Sql().Create(&catHigh).Error; err != nil {
		t.Fatalf("create high sort category: %v", err)
	}

	for _, catID := range []uint64{catLow.ID, catHigh.ID} {
		link := database.PostCategory{PostID: post.ID, CategoryID: catID}
		if err := conn.Sql().Create(&link).Error; err != nil {
			t.Fatalf("create link: %v", err)
		}
	}

	h := NewCategoriesHandler(&repository.Categories{DB: conn})
	req := httptest.NewRequest("GET", "/categories", nil)
	rec := httptest.NewRecorder()

	if err := h.Index(rec, req); err != nil {
		t.Fatalf("index err: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var resp pagination.Pagination[payload.CategoryResponse]

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Data) != 2 {
		t.Fatalf("unexpected data length: %+v", resp.Data)
	}

	if resp.Data[0].Name != "Zebra" || resp.Data[0].Sort != 5 {
		t.Fatalf("unexpected first category: %+v", resp.Data[0])
	}

	if resp.Data[1].Name != "Apple" || resp.Data[1].Sort != 20 {
		t.Fatalf("unexpected second category: %+v", resp.Data[1])
	}
}
