package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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
	conn, author := handlertests.MakeTestDB(t)

	type fixture struct {
		Post struct {
			Slug    string `json:"slug"`
			Title   string `json:"title"`
			Excerpt string `json:"excerpt"`
			Content string `json:"content"`
		} `json:"post"`
		Category struct {
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Description string `json:"description"`
		} `json:"category"`
	}

	f, err := os.Open("../storage/fixture/categories.json")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	var fx fixture

	if err := json.NewDecoder(f).Decode(&fx); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	published := time.Now()

	post := database.Post{
		UUID:        uuid.NewString(),
		AuthorID:    author.ID,
		Slug:        fx.Post.Slug,
		Title:       fx.Post.Title,
		Excerpt:     fx.Post.Excerpt,
		Content:     fx.Post.Content,
		PublishedAt: &published,
	}

	if err := conn.Sql().Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	cat := database.Category{
		UUID:        uuid.NewString(),
		Name:        fx.Category.Name,
		Slug:        fx.Category.Slug,
		Description: fx.Category.Description,
	}

	if err := conn.Sql().Create(&cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	link := database.PostCategory{
		PostID:     post.ID,
		CategoryID: cat.ID,
	}

	if err := conn.Sql().Create(&link).Error; err != nil {
		t.Fatalf("create link: %v", err)
	}

	h := MakeCategoriesHandler(&repository.Categories{
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

	if len(resp.Data) != 1 || resp.Data[0].Slug != "cat" {
		t.Fatalf("unexpected data: %+v", resp.Data)
	}
}
