package support

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
	"github.com/oullin/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/oullin/metal/env"
)

// TestsHelper provides common test utilities for database-backed tests.
type TestsHelper struct {
	t    *testing.T
	conn *database.Connection
	env  *env.Environment
}

// NewTestsHelper creates a new test helper with a Postgres connection.
// The models parameter accepts database models to auto-migrate.
func NewTestsHelper(t *testing.T, models ...interface{}) *TestsHelper {
	t.Helper()

	conn, e := newPostgresConnection(t, models...)

	return &TestsHelper{
		t:    t,
		conn: conn,
		env:  e,
	}
}

// Conn returns the database connection.
func (h *TestsHelper) Conn() *database.Connection {
	return h.conn
}

// Env returns the environment configuration.
func (h *TestsHelper) Env() *env.Environment {
	return h.env
}

// SeedCategory creates and persists a category for testing.
func (h *TestsHelper) SeedCategory(slug, name string, sort int) database.Category {
	h.t.Helper()

	category := database.Category{
		UUID: uuid.NewString(),
		Slug: slug,
		Name: name,
		Sort: sort,
	}

	if err := h.conn.Sql().Create(&category).Error; err != nil {
		h.t.Fatalf("create category: %v", err)
	}

	return category
}

// SeedTag creates and persists a tag for testing.
func (h *TestsHelper) SeedTag(slug, name string) database.Tag {
	h.t.Helper()

	tag := database.Tag{
		UUID: uuid.NewString(),
		Slug: slug,
		Name: name,
	}

	if err := h.conn.Sql().Create(&tag).Error; err != nil {
		h.t.Fatalf("create tag: %v", err)
	}

	return tag
}

// SeedUser creates and persists a user for testing.
func (h *TestsHelper) SeedUser(first, last, username string) database.User {
	h.t.Helper()

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

	if err := h.conn.Sql().Create(&user).Error; err != nil {
		h.t.Fatalf("create user: %v", err)
	}

	return user
}

// SeedPost creates and persists a post for testing.
// The published parameter determines if the post should have a published timestamp.
func (h *TestsHelper) SeedPost(author database.User, category database.Category, tag database.Tag, slug, title string, published bool) database.Post {
	h.t.Helper()

	var publishedAt *time.Time
	if published {
		ts := time.Now().UTC()
		publishedAt = &ts
	}

	post := database.Post{
		UUID:        uuid.NewString(),
		AuthorID:    author.ID,
		Slug:        slug,
		Title:       title,
		Excerpt:     title + " excerpt",
		Content:     title + " content",
		PublishedAt: publishedAt,
	}

	if err := h.conn.Sql().Create(&post).Error; err != nil {
		h.t.Fatalf("create post: %v", err)
	}

	// Create category association
	postCategory := database.PostCategory{
		PostID:     post.ID,
		CategoryID: category.ID,
	}
	if err := h.conn.Sql().Create(&postCategory).Error; err != nil {
		h.t.Fatalf("create post category: %v", err)
	}

	// Create tag association
	postTag := database.PostTag{
		PostID: post.ID,
		TagID:  tag.ID,
	}
	if err := h.conn.Sql().Create(&postTag).Error; err != nil {
		h.t.Fatalf("create post tag: %v", err)
	}

	// Load associations
	if err := h.conn.Sql().Preload("Categories").Preload("Tags").Preload("Author").First(&post, post.ID).Error; err != nil {
		h.t.Fatalf("load post associations: %v", err)
	}

	return post
}

// SeedPostWithContent creates and persists a post with custom content for testing.
// This is useful for SEO tests that need specific HTML content.
func (h *TestsHelper) SeedPostWithContent(author database.User, category database.Category, tag database.Tag, slug, title, excerpt, content, imageURL string) database.Post {
	h.t.Helper()

	publishedAt := time.Now().UTC()

	post := database.Post{
		UUID:          uuid.NewString(),
		AuthorID:      author.ID,
		Slug:          slug,
		Title:         title,
		Excerpt:       excerpt,
		Content:       content,
		CoverImageURL: imageURL,
		PublishedAt:   &publishedAt,
	}

	if err := h.conn.Sql().Create(&post).Error; err != nil {
		h.t.Fatalf("create post: %v", err)
	}

	// Create category association
	postCategory := database.PostCategory{
		PostID:     post.ID,
		CategoryID: category.ID,
	}
	if err := h.conn.Sql().Create(&postCategory).Error; err != nil {
		h.t.Fatalf("create post category: %v", err)
	}

	// Create tag association
	postTag := database.PostTag{
		PostID: post.ID,
		TagID:  tag.ID,
	}
	if err := h.conn.Sql().Create(&postTag).Error; err != nil {
		h.t.Fatalf("create post tag: %v", err)
	}

	// Load associations
	if err := h.conn.Sql().Preload("Categories").Preload("Tags").Preload("Author").First(&post, post.ID).Error; err != nil {
		h.t.Fatalf("load post associations: %v", err)
	}

	return post
}

// newPostgresConnection creates a new Postgres test container and database connection.
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
			Name:      "Test App",
			URL:       "https://test.example.test",
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

// WithRepoRoot changes the working directory to the repository root for the duration of the test.
// This is useful for tests that need to access files relative to the repo root.
func WithRepoRoot(t *testing.T) {
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
