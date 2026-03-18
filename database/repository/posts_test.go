package repository_test

import (
	"testing"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/database/repository/queries"
	"github.com/oullin/internal/testutil/dbtest"
)

func TestPostsCreateLinksAssociationsPostgres(t *testing.T) {
	h := dbtest.NewTestsHelper(t,
		&database.User{},
		&database.Post{},
		&database.Category{},
		&database.PostCategory{},
		&database.Tag{},
		&database.PostTag{},
	)

	user := h.SeedUser("Alice", "Smith", "alice")
	category := h.SeedCategory("tech", "Tech", 1)
	tag := h.SeedTag("go", "Go")

	conn := h.Conn()

	postsRepo := repository.Posts{
		DB:         conn,
		Categories: &repository.Categories{DB: conn},
		Tags:       &repository.Tags{DB: conn},
	}

	publishedAt := time.Now().UTC()

	post, err := postsRepo.Create(database.PostsAttrs{
		AuthorID:    user.ID,
		Slug:        "first-post",
		Title:       "First Post",
		Excerpt:     "First excerpt",
		Content:     "First content",
		ImageURL:    "https://example.test/cover.png",
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

	if post.ID == 0 || post.UUID == "" {
		t.Fatalf("expected persisted post with identifiers")
	}

	var categoryLinks int64
	if err := conn.Sql().Model(&database.PostCategory{}).Where("post_id = ?", post.ID).Count(&categoryLinks).Error; err != nil {
		t.Fatalf("count post categories: %v", err)
	}

	if categoryLinks != 1 {
		t.Fatalf("expected 1 category link, got %d", categoryLinks)
	}

	var tagLinks int64
	if err := conn.Sql().Model(&database.PostTag{}).Where("post_id = ?", post.ID).Count(&tagLinks).Error; err != nil {
		t.Fatalf("count post tags: %v", err)
	}

	if tagLinks != 1 {
		t.Fatalf("expected 1 tag link, got %d", tagLinks)
	}
}

func TestPostsFindByLoadsAssociationsPostgres(t *testing.T) {
	h := dbtest.NewTestsHelper(t,
		&database.User{},
		&database.Post{},
		&database.Category{},
		&database.PostCategory{},
		&database.Tag{},
		&database.PostTag{},
	)

	user := h.SeedUser("Bob", "Jones", "bobj")
	category := h.SeedCategory("career", "Career", 1)
	tag := h.SeedTag("work", "Work")
	post := h.SeedPost(user, category, tag, "career-path", "Career Path", true)

	conn := h.Conn()

	postsRepo := repository.Posts{
		DB:         conn,
		Categories: &repository.Categories{DB: conn},
		Tags:       &repository.Tags{DB: conn},
	}

	found := postsRepo.FindBy("career-path")
	if found == nil {
		t.Fatalf("expected to find post")
	}

	if found.ID != post.ID {
		t.Fatalf("expected matching post id")
	}

	if len(found.Categories) != 1 || found.Categories[0].ID != category.ID {
		t.Fatalf("expected category association to load")
	}

	if len(found.Tags) != 1 || found.Tags[0].ID != tag.ID {
		t.Fatalf("expected tag association to load")
	}

	if found.Author.ID != user.ID {
		t.Fatalf("expected author association to load")
	}

	if postsRepo.FindBy("missing") != nil {
		t.Fatalf("expected missing post lookup to return nil")
	}
}

func TestPostsGetAllFiltersPublishedRecordsPostgres(t *testing.T) {
	h := dbtest.NewTestsHelper(t,
		&database.User{},
		&database.Post{},
		&database.Category{},
		&database.PostCategory{},
		&database.Tag{},
		&database.PostTag{},
	)

	authorOne := h.SeedUser("Carol", "One", "carol")
	authorTwo := h.SeedUser("Dave", "Two", "dave")

	category := h.SeedCategory("engineering", "Engineering", 1)
	tag := h.SeedTag("backend", "Backend")
	otherTag := h.SeedTag("frontend", "Frontend")

	published := h.SeedPost(authorOne, category, tag, "backend-guide", "Backend Guide", true)
	deleted := h.SeedPost(authorTwo, category, otherTag, "frontend-guide", "Frontend Guide", true)
	_ = h.SeedPost(authorTwo, category, otherTag, "draft-guide", "Draft Guide", false)

	conn := h.Conn()

	if err := conn.Sql().Delete(&database.Post{}, deleted.ID).Error; err != nil {
		t.Fatalf("soft delete post: %v", err)
	}

	postsRepo := repository.Posts{DB: conn}

	paginate := pagination.Paginate{Page: 1, Limit: 5}

	result, err := postsRepo.GetAll(queries.PostFilters{}, paginate)
	if err != nil {
		t.Fatalf("get all: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected single result, got %d", len(result.Data))
	}

	if result.Data[0].Slug != published.Slug {
		t.Fatalf("expected only published post, got %q", result.Data[0].Slug)
	}

	if result.PageSize != 5 || result.Page != 1 {
		t.Fatalf("unexpected pagination metadata: %+v", result)
	}

	if result.NextPage != nil {
		t.Fatalf("expected no next page for single result")
	}

	if result.PreviousPage != nil {
		t.Fatalf("expected no previous page for first page")
	}
}

func TestPostsGetAllDeduplicatesResultsPostgres(t *testing.T) {
	h := dbtest.NewTestsHelper(t,
		&database.User{},
		&database.Post{},
		&database.Category{},
		&database.PostCategory{},
		&database.Tag{},
		&database.PostTag{},
	)

	author := h.SeedUser("Eve", "Duplicates", "eve")

	primaryCategory := h.SeedCategory("engineering", "Engineering", 1)
	secondaryCategory := h.SeedCategory("engagement", "Engagement", 2)

	primaryTag := h.SeedTag("eng-backend", "Eng Backend")
	secondaryTag := h.SeedTag("eng-frontend", "Eng Frontend")

	post := h.SeedPost(author, primaryCategory, primaryTag, "dedupe-check", "Dedupe Check", true)

	conn := h.Conn()

	extraCategory := database.PostCategory{PostID: post.ID, CategoryID: secondaryCategory.ID}
	if err := conn.Sql().Create(&extraCategory).Error; err != nil {
		t.Fatalf("create secondary category relation: %v", err)
	}

	extraTag := database.PostTag{PostID: post.ID, TagID: secondaryTag.ID}
	if err := conn.Sql().Create(&extraTag).Error; err != nil {
		t.Fatalf("create secondary tag relation: %v", err)
	}

	postsRepo := repository.Posts{DB: conn}

	filters := queries.PostFilters{Category: "eng", Tag: "eng"}
	paginate := pagination.Paginate{Page: 1, Limit: 5}

	result, err := postsRepo.GetAll(filters, paginate)
	if err != nil {
		t.Fatalf("get all: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	if len(result.Data) != 1 {
		t.Fatalf("expected single result, got %d", len(result.Data))
	}

	if result.Data[0].ID != post.ID {
		t.Fatalf("expected post %d, got %d", post.ID, result.Data[0].ID)
	}
}

func TestPostsFindCategoryByDelegatesPostgres(t *testing.T) {
	h := dbtest.NewTestsHelper(t, &database.Category{})

	category := h.SeedCategory("lifestyle", "Lifestyle", 1)

	conn := h.Conn()

	postsRepo := repository.Posts{
		DB:         conn,
		Categories: &repository.Categories{DB: conn},
	}

	if found := postsRepo.FindCategoryBy("LIFESTYLE"); found == nil || found.ID != category.ID {
		t.Fatalf("expected category lookup to delegate to categories repository")
	}
}

func TestPostsFindTagByHandlesRepositoryErrorsPostgres(t *testing.T) {
	h := dbtest.NewTestsHelper(t, &database.Tag{})

	conn := h.Conn()

	sqlDB, err := conn.Sql().DB()
	if err != nil {
		t.Fatalf("unwrap sql db: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}

	postsRepo := repository.Posts{
		DB:   conn,
		Tags: &repository.Tags{DB: conn},
	}

	if tag := postsRepo.FindTagBy("anything"); tag != nil {
		t.Fatalf("expected nil tag when repository errors")
	}
}
