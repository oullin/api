package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/env"
	"github.com/oullin/handler/payload"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func makeCategoriesRepo(t *testing.T) *repository.Categories {
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

	if err := conn.Sql().AutoMigrate(&database.User{}, &database.Post{}, &database.Category{}, &database.PostCategory{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	author := database.User{ID: 1, UUID: uuid.NewString(), Username: "user", FirstName: "F", LastName: "L", Email: "u@example.com", PasswordHash: "x"}
	if err := conn.Sql().Create(&author).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	published := time.Now()
	post := database.Post{UUID: uuid.NewString(), AuthorID: author.ID, Slug: "hello", Title: "Hello", Excerpt: "Ex", Content: "Body", PublishedAt: &published}
	if err := conn.Sql().Create(&post).Error; err != nil {
		t.Fatalf("create post: %v", err)
	}

	cat := database.Category{UUID: uuid.NewString(), Name: "Cat", Slug: "cat", Description: "desc"}
	if err := conn.Sql().Create(&cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	link := database.PostCategory{PostID: post.ID, CategoryID: cat.ID}
	if err := conn.Sql().Create(&link).Error; err != nil {
		t.Fatalf("create link: %v", err)
	}

	return &repository.Categories{DB: conn}
}

func TestCategoriesHandlerIndex_Success(t *testing.T) {
	repo := makeCategoriesRepo(t)
	h := MakeCategoriesHandler(repo)

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
