package seeds

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/oullin/database"
	"github.com/oullin/metal/env"
)

func testConnection(t *testing.T, e *env.Environment) *database.Connection {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
	if err := exec.Command("docker", "ps").Run(); err != nil {
		t.Skip("docker not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	t.Cleanup(func() { pg.Terminate(context.Background()) })

	host, err := pg.Host(ctx)

	if err != nil {
		t.Fatalf("host err: %v", err)
	}

	port, err := pg.MappedPort(ctx, "5432/tcp")

	if err != nil {
		t.Fatalf("port err: %v", err)
	}

	e.DB = env.DBEnvironment{
		UserName:     "test",
		UserPassword: "secret",
		DatabaseName: "testdb",
		Port:         port.Int(),
		Host:         host,
		DriverName:   database.DriverName,
		SSLMode:      "disable",
		TimeZone:     "UTC",
	}

	conn, err := database.NewConnection(e)

	if err != nil {
		t.Fatalf("make connection: %v", err)
	}

	t.Cleanup(func() { conn.Close() })

	if err := conn.Sql().AutoMigrate(
		&database.User{},
		&database.Post{},
		&database.Category{},
		&database.PostCategory{},
		&database.Tag{},
		&database.PostTag{},
		&database.PostView{},
		&database.Comment{},
		&database.Like{},
		&database.Newsletter{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return conn
}

func setupSeeder(t *testing.T) *Seeder {
	e := &env.Environment{App: env.AppEnvironment{Type: "local"}}

	conn := testConnection(t, e)

	return NewSeeder(conn, e)
}

func TestSeederWorkflow(t *testing.T) {
	seeder := setupSeeder(t)

	if err := seeder.TruncateDB(); err != nil {
		t.Fatalf("truncate err: %v", err)
	}

	userA, userB := seeder.SeedUsers()
	posts := seeder.SeedPosts(userA, userB)
	categories := seeder.SeedCategories()
	tags := seeder.SeedTags()
	seeder.SeedComments(posts...)
	seeder.SeedLikes(posts...)
	seeder.SeedPostsCategories(categories, posts)
	seeder.SeedPostTags(tags, posts)
	seeder.SeedPostViews(posts, userA, userB)

	if err := seeder.SeedNewsLetters(); err != nil {
		t.Fatalf("newsletter err: %v", err)
	}

	var count int64

	seeder.dbConn.Sql().Model(&database.User{}).Count(&count)

	if count != 2 {
		t.Fatalf("expected 2 users got %d", count)
	}

	seeder.dbConn.Sql().Model(&database.Post{}).Count(&count)

	if count != 2 {
		t.Fatalf("expected 2 posts got %d", count)
	}

	seeder.dbConn.Sql().Model(&database.Category{}).Count(&count)

	if count == 0 {
		t.Fatalf("categories not seeded")
	}
}

func TestSeederEmptyMethods(t *testing.T) {
	seeder := setupSeeder(t)

	seeder.SeedPostsCategories(nil, nil)
	seeder.SeedPostTags(nil, nil)
	seeder.SeedPostViews(nil)

	var count int64

	seeder.dbConn.Sql().Model(&database.PostCategory{}).Count(&count)

	if count != 0 {
		t.Fatalf("expected 0 post_categories")
	}

	seeder.dbConn.Sql().Model(&database.PostTag{}).Count(&count)

	if count != 0 {
		t.Fatalf("expected 0 post_tags")
	}

	seeder.dbConn.Sql().Model(&database.PostView{}).Count(&count)

	if count != 0 {
		t.Fatalf("expected 0 post_views")
	}
}
