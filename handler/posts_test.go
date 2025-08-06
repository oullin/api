package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/handler/payload"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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

func makePostsRepo(t *testing.T) *repository.Posts {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.AutoMigrate(&database.User{}, &database.Post{}, &database.Category{}, &database.Tag{}, &database.PostCategory{}, &database.PostTag{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	conn := &database.Connection{}
	rv := reflect.ValueOf(conn).Elem()
	driverField := rv.FieldByName("driver")
	reflect.NewAt(driverField.Type(), unsafe.Pointer(driverField.UnsafeAddr())).Elem().Set(reflect.ValueOf(db))
	nameField := rv.FieldByName("driverName")
	reflect.NewAt(nameField.Type(), unsafe.Pointer(nameField.UnsafeAddr())).Elem().SetString("sqlite")

	author := database.User{ID: 1, UUID: "u1", Username: "user", FirstName: "F", LastName: "L", Email: "u@example.com", PasswordHash: "x"}
	if err := db.Create(&author).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	published := time.Now()
	post := database.Post{UUID: "p1", AuthorID: author.ID, Slug: "hello", Title: "Hello", Excerpt: "Ex", Content: "Body", PublishedAt: &published}
	if err := db.Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	return &repository.Posts{DB: conn}
}

func TestPostsHandlerIndex_Success(t *testing.T) {
	repo := makePostsRepo(t)
	h := MakePostsHandler(repo)

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
	repo := makePostsRepo(t)
	h := MakePostsHandler(repo)

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
}
