package repository_test

import (
	"testing"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
)

func TestUsersFindBy(t *testing.T) {
	conn := newPostgresConnection(t, &database.User{})

	user := seedUser(t, conn, "John", "Doe", "jdoe")

	repo := repository.Users{DB: conn}

	found := repo.FindBy("jdoe")

	if found == nil || found.ID != user.ID {
		t.Fatalf("user not found")
	}
}

func TestTagsFindOrCreate(t *testing.T) {
	conn := newPostgresConnection(t, &database.Tag{})

	repo := repository.Tags{DB: conn}

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
	conn := newPostgresConnection(t, &database.Category{})

	category := seedCategory(t, conn, "news", "News", 1)

	repo := repository.Categories{DB: conn}

	found := repo.FindBy("news")

	if found == nil || found.ID != category.ID {
		t.Fatalf("category not found")
	}
}

func TestPostsCreateAndFind(t *testing.T) {
	conn := newPostgresConnection(t,
		&database.User{},
		&database.Post{},
		&database.Category{},
		&database.PostCategory{},
		&database.Tag{},
		&database.PostTag{},
	)

	user := seedUser(t, conn, "Jane", "Doe", "jane")
	category := seedCategory(t, conn, "tech", "Tech", 1)
	tag := seedTag(t, conn, "go", "Go")

	postsRepo := repository.Posts{
		DB:         conn,
		Categories: &repository.Categories{DB: conn},
		Tags:       &repository.Tags{DB: conn},
	}

	post, err := postsRepo.Create(database.PostsAttrs{
		AuthorID: user.ID,
		Slug:     "first-post",
		Title:    "First Post",
		Excerpt:  "excerpt",
		Content:  "content",
		Categories: []database.CategoriesAttrs{
			{
				Id:   category.ID,
				Name: category.Name,
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

	if found == nil || found.ID != post.ID {
		t.Fatalf("post not found")
	}
}
