package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/env"
	"github.com/oullin/handler/payload"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
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

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}

	ctx := context.Background()
	pg, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("secret"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("container run err: %v", err)
	}
	t.Cleanup(func() { pg.Terminate(ctx) })

	host, err := pg.Host(ctx)
	if err != nil {
		t.Fatalf("host err: %v", err)
	}
	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("port err: %v", err)
	}

	e := &env.Environment{
		DB: env.DBEnvironment{
			UserName:     "test",
			UserPassword: "secret",
			DatabaseName: "testdb",
			Port:         port.Int(),
			Host:         host,
			DriverName:   database.DriverName,
			SSLMode:      "disable",
			TimeZone:     "UTC",
		},
	}

	conn, err := database.MakeConnection(e)
	if err != nil {
		t.Fatalf("make connection: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := conn.Sql().AutoMigrate(&database.User{}, &database.Post{}, &database.Category{}, &database.Tag{}, &database.PostCategory{}, &database.PostTag{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	author := database.User{ID: 1, UUID: "u1", Username: "user", FirstName: "F", LastName: "L", Email: "u@example.com", PasswordHash: "x"}
	if err := conn.Sql().Create(&author).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	published := time.Now()
	post := database.Post{UUID: "p1", AuthorID: author.ID, Slug: "hello", Title: "Hello", Excerpt: "Ex", Content: "Body", PublishedAt: &published}
	if err := conn.Sql().Create(&post).Error; err != nil {
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
