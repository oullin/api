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

func setupDB(t *testing.T) *database.Connection {
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

	return conn
}

func TestUsersFindBy(t *testing.T) {
	conn := setupDB(t)

	if err := conn.Sql().AutoMigrate(&database.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	u := database.User{
		UUID:         uuid.NewString(),
		FirstName:    "John",
		LastName:     "Doe",
		Username:     "jdoe",
		DisplayName:  "John Doe",
		Email:        "jdoe@test.com",
		PasswordHash: "x",
		PublicToken:  uuid.NewString(),
	}

	if err := conn.Sql().Create(&u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	repo := repository.Users{
		DB:  conn,
		Env: &env.Environment{},
	}

	found := repo.FindBy("jdoe")

	if found == nil || found.ID != u.ID {
		t.Fatalf("user not found")
	}
}

func TestTagsFindOrCreate(t *testing.T) {
	conn := setupDB(t)

	if err := conn.Sql().AutoMigrate(&database.Tag{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := repository.Tags{
		DB: conn,
	}

	first, err := repo.FindOrCreate("golang")

	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	second, err := repo.FindOrCreate("golang")

	if err != nil {
		t.Fatalf("find tag: %v", err)
	}

	if first.ID != second.ID {
		t.Fatalf("expected same tag")
	}
}

func TestCategoriesFindBy(t *testing.T) {
	conn := setupDB(t)

	if err := conn.Sql().AutoMigrate(&database.Category{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	c := database.Category{
		UUID: uuid.NewString(),
		Name: "News",
		Slug: "news",
	}

	if err := conn.Sql().Create(&c).Error; err != nil {
		t.Fatalf("create cat: %v", err)
	}

	repo := repository.Categories{
		DB: conn,
	}

	found := repo.FindBy("news")

	if found == nil || found.ID != c.ID {
		t.Fatalf("category not found")
	}
}

func TestPostsCreateAndFind(t *testing.T) {
	conn := setupDB(t)

	if err := conn.Sql().AutoMigrate(&database.User{}, &database.Post{}, &database.Category{}, &database.PostCategory{}, &database.Tag{}, &database.PostTag{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	user := database.User{
		UUID:         uuid.NewString(),
		FirstName:    "Jane",
		LastName:     "Doe",
		Username:     "jane",
		DisplayName:  "Jane Doe",
		Email:        "jane@test.com",
		PasswordHash: "x",
		PublicToken:  uuid.NewString(),
	}

	if err := conn.Sql().Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	cat := database.Category{
		UUID: uuid.NewString(),
		Name: "Tech",
		Slug: "tech",
	}

	if err := conn.Sql().Create(&cat).Error; err != nil {
		t.Fatalf("create cat: %v", err)
	}

	tag := database.Tag{
		UUID: uuid.NewString(),
		Name: "Go",
		Slug: "go",
	}

	if err := conn.Sql().Create(&tag).Error; err != nil {
		t.Fatalf("create tag: %v", err)
	}

	postsRepo := repository.Posts{
		DB: conn,
		Categories: &repository.Categories{
			DB: conn,
		},
		Tags: &repository.Tags{
			DB: conn,
		},
	}

	p, err := postsRepo.Create(database.PostsAttrs{
		AuthorID: user.ID,
		Slug:     "first-post",
		Title:    "First Post",
		Excerpt:  "excerpt",
		Content:  "content",
		Categories: []database.CategoriesAttrs{
			{
				Id:   cat.ID,
				Name: cat.Name,
			},
		},
		Tags: []database.TagAttrs{
			{
				Id:   tag.ID,
				Name: tag.Name,
			},
		},
	})

	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	found := postsRepo.FindBy("first-post")

	if found == nil || found.ID != p.ID {
		t.Fatalf("post not found")
	}
}
