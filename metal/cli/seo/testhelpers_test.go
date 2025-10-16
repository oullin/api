package seo

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
)

func newPostgresConnection(t *testing.T, models ...interface{}) (*database.Connection, *env.Environment) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}

	if err := exec.Command("docker", "ps").Run(); err != nil {
		t.Skip("docker not running")
	}

	const (
		username = "testaccount"
		password = "secretpassw"
		dbname   = "testdb"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)

	pg, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		postgres.WithDatabase(dbname),
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

	spaDir := t.TempDir()

	e := &env.Environment{
		App: env.AppEnvironment{
			Name:      "SEO Test App",
			URL:       "https://seo.example.test",
			Type:      "local",
			MasterKey: strings.Repeat("m", 32),
		},
		DB: env.DBEnvironment{
			UserName:     username,
			UserPassword: password,
			DatabaseName: dbname,
			Port:         port.Int(),
			Host:         host,
			DriverName:   database.DriverName,
			SSLMode:      "disable",
			TimeZone:     "UTC",
		},
		Logs: env.LogsEnvironment{
			Level:      "info",
			Dir:        "logs",
			DateFormat: "yyyy-mm",
		},
		Network: env.NetEnvironment{
			HttpHost: "localhost",
			HttpPort: "8080",
		},
		Sentry: env.SentryEnvironment{
			DSN: "dsn",
			CSP: "csp",
		},
		Ping: env.PingEnvironment{
			Username: strings.Repeat("p", 16),
			Password: strings.Repeat("s", 16),
		},
		Seo: env.SeoEnvironment{
			SpaDir:       spaDir,
			SpaImagesDir: filepath.Join(spaDir, "posts", "images"),
		},
	}

	conn, err := database.NewConnection(e)
	if err != nil {
		t.Fatalf("new connection: %v", err)
	}

	if len(models) > 0 {
		if err := conn.Sql().AutoMigrate(models...); err != nil {
			t.Fatalf("auto migrate: %v", err)
		}
	}

	t.Cleanup(func() {
		if err := conn.Ping(); err == nil {
			conn.Close()
		}

		_ = pg.Terminate(context.Background())
	})

	return conn, e
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

func seedUser(t *testing.T, conn *database.Connection, first, last, username string) database.User {
	t.Helper()

	var parts []string
	if trimmed := strings.TrimSpace(first); trimmed != "" {
		parts = append(parts, trimmed)
	}
	if trimmed := strings.TrimSpace(last); trimmed != "" {
		parts = append(parts, trimmed)
	}
	display := strings.Join(parts, " ")
	if display == "" {
		display = username
	}

	user := database.User{
		UUID:         uuid.NewString(),
		FirstName:    first,
		LastName:     last,
		Username:     username,
		DisplayName:  display,
		Email:        fmt.Sprintf("%s@example.test", strings.TrimSpace(username)),
		PasswordHash: strings.Repeat("p", 60),
		PublicToken:  uuid.NewString(),
		VerifiedAt:   time.Now().UTC(),
	}

	if err := conn.Sql().Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user
}

func seedPost(t *testing.T, conn *database.Connection, author database.User, category database.Category, tag database.Tag, slug, title string) database.Post {
	t.Helper()

	postsRepo := repository.Posts{
		DB:         conn,
		Categories: &repository.Categories{DB: conn},
		Tags:       &repository.Tags{DB: conn},
	}

	publishedAt := time.Now().UTC()

	post, err := postsRepo.Create(database.PostsAttrs{
		AuthorID: author.ID,
		Slug:     slug,
		Title:    title,
		// Seed values include characters that should be HTML-escaped so
		// integration tests can assert the generated SEO pages are safe.
		Excerpt:     "Learn <fast>\nwith examples",
		Content:     "Intro paragraph with <tags>\nmore info.\n\nSecond paragraph & details.",
		ImageURL:    fmt.Sprintf("https://seo.example.test/%s.png", slug),
		PublishedAt: &publishedAt,
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

func withRepoRoot(t *testing.T) {
	t.Helper()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	root := filepath.Clean(filepath.Join(cwd, "..", "..", ".."))

	if err := os.Chdir(root); err != nil {
		t.Fatalf("change to repo root: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chdir(cwd)
	})
}
