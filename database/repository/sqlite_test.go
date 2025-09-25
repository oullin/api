package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
)

func newSQLiteConnection(t *testing.T) (*database.Connection, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("unwrap sql db: %v", err)
	}

	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return database.NewConnectionFromGorm(db), db
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
