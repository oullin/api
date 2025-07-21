package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository/queries"
	"github.com/oullin/pkg/gorm"
)

type Posts struct {
	DB         *database.Connection
	Categories *Categories
	Tags       *Tags
}

func (p Posts) GetPosts(filters *queries.PostFilters, pagination *Pagination[database.Post]) (*Pagination[database.Post], error) {
	page := 1
	pageSize := 10
	var total int64
	var posts []database.Post

	query := p.
		DB.Sql().
		Model(&database.Post{}).
		Distinct("posts.id, posts.published_at")

	queries.ApplyPostsFilters(filters, query)

	if pagination != nil {
		page = pagination.Page
		pageSize = pagination.PageSize
	}

	countQuery := query.Session(p.DB.Session())
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Fetch the data for the current page
	offset := (page - 1) * pageSize
	err := query.Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Order("posts.published_at DESC").
		Limit(pageSize).
		Offset(offset).
		Distinct().
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	result := MakePagination[database.Post](posts, page, pageSize, total)

	return result, nil
}

func (p Posts) FindCategoryBy(slug string) *database.Category {
	return p.Categories.FindBy(slug)
}

func (p Posts) FindTagBy(slug string) *database.Tag {
	tag, err := p.Tags.FindOrCreate(slug)

	if err != nil {
		return nil
	}

	return tag
}

func (p Posts) Create(attrs database.PostsAttrs) (*database.Post, error) {
	post := database.Post{
		UUID:          uuid.NewString(),
		AuthorID:      attrs.AuthorID,
		Slug:          attrs.Slug,
		Title:         attrs.Title,
		Excerpt:       attrs.Excerpt,
		Content:       attrs.Content,
		CoverImageURL: attrs.ImageURL,
		PublishedAt:   attrs.PublishedAt,
	}

	if result := p.DB.Sql().Create(&post); gorm.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("issue creating posts: %s", result.Error)
	}

	if err := p.LinkCategories(post, attrs.Categories); err != nil {
		return nil, fmt.Errorf("issue creating the given post [%s] category: %s", attrs.Slug, err.Error())
	}

	if err := p.LinkTags(post, attrs.Tags); err != nil {
		return nil, fmt.Errorf("issue creating the given post [%s] tags: %s", attrs.Slug, err.Error())
	}

	return &post, nil
}

func (p Posts) LinkCategories(post database.Post, categories []database.CategoriesAttrs) error {
	for _, category := range categories {
		trace := database.PostCategory{
			CategoryID: category.Id,
			PostID:     post.ID,
		}

		if result := p.DB.Sql().Create(&trace); gorm.HasDbIssues(result.Error) {
			return fmt.Errorf("error linking categories [%s:%s]: %s", category.Name, post.Title, result.Error)
		}
	}

	return nil
}

func (p Posts) LinkTags(post database.Post, tags []database.TagAttrs) error {
	for _, tag := range tags {
		trace := database.PostTag{
			TagID:  tag.Id,
			PostID: post.ID,
		}

		if result := p.DB.Sql().Create(&trace); gorm.HasDbIssues(result.Error) {
			return fmt.Errorf("error linking tags [%s:%s]: %s", tag.Name, post.Title, result.Error)
		}
	}

	return nil
}
