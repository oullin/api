package repository_test

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
)

func newPostgresConnection(t *testing.T, models ...interface{}) *database.Connection {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}

	if err := exec.Command("docker", "ps").Run(); err != nil {
		t.Skip("docker not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

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

	if len(models) > 0 {
		if err := conn.Sql().AutoMigrate(models...); err != nil {
			t.Fatalf("migrate schema: %v", err)
		}
	}

	t.Cleanup(func() {
		if err := conn.Ping(); err == nil {
			conn.Close()
		}

		_ = pg.Terminate(context.Background())
	})

	return conn
}

func seedUser(t *testing.T, conn *database.Connection, first, last, username string) database.User {
	t.Helper()

	user := database.User{
		UUID:         uuid.NewString(),
		FirstName:    first,
		LastName:     last,
		Username:     username,
		DisplayName:  first + " " + last,
		Email:        username + "@example.test",
		PasswordHash: "hash",
		PublicToken:  uuid.NewString(),
	}

	if err := conn.Sql().Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user
}

func seedCategory(t *testing.T, conn *database.Connection, slug, name string) database.Category {
	t.Helper()

	category := database.Category{
		UUID: uuid.NewString(),
		Slug: slug,
		Name: name,
	}

	if err := conn.Sql().Create(&category).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	return category
}

func seedTag(t *testing.T, conn *database.Connection, slug, name string) database.Tag {
	t.Helper()

	tag := database.Tag{
		UUID: uuid.NewString(),
		Slug: slug,
		Name: name,
	}

	if err := conn.Sql().Create(&tag).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}

	return tag
}

func seedPost(t *testing.T, conn *database.Connection, author database.User, category database.Category, tag database.Tag, slug, title string, published bool) database.Post {
	t.Helper()

	postsRepo := repository.Posts{
		DB:         conn,
		Categories: &repository.Categories{DB: conn},
		Tags:       &repository.Tags{DB: conn},
	}

	var publishedAt *time.Time
	if published {
		ts := time.Now().UTC()
		publishedAt = &ts
	}

	post, err := postsRepo.Create(database.PostsAttrs{
		AuthorID:    author.ID,
		Slug:        slug,
		Title:       title,
		Excerpt:     title + " excerpt",
		Content:     title + " content",
		PublishedAt: publishedAt,
		Categories: []database.CategoriesAttrs{{
			Id:   category.ID,
			Slug: category.Slug,
			Name: category.Name,
		}},
		Tags: []database.TagAttrs{{
			Id:   tag.ID,
			Slug: tag.Slug,
			Name: tag.Name,
		}},
	})
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	return *post
}
