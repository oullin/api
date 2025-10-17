package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/database/repository/queries"
	"github.com/oullin/pkg/model"
)

type Posts struct {
	DB         *database.Connection
	Categories *Categories
	Tags       *Tags
}

func (p Posts) GetAll(filters queries.PostFilters, paginate pagination.Paginate) (*pagination.Pagination[database.Post], error) {
	var numItems int64
	var posts []database.Post

	query := p.DB.Sql().
		Model(&database.Post{}).
		Where("posts.published_at is not null"). // only published posts will be selected.
		Where("posts.deleted_at is null")        // deleted posted will be discarded.

	queries.ApplyPostsFilters(&filters, query)

	if err := pagination.Count[*int64](&numItems, query, p.DB.GetSession(), "posts.id"); err != nil {
		return nil, err
	}

	offset := (paginate.Page - 1) * paginate.Limit

	err := query.Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Order("posts.published_at DESC, posts.id DESC").
		Select("DISTINCT ON (posts.published_at, posts.id) posts.*"). // ensure joined relations do not duplicate posts
		Limit(paginate.Limit).
		Offset(offset).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	paginate.SetNumItems(numItems)
	result := pagination.NewPagination[database.Post](posts, paginate)

	return result, nil
}

func (p Posts) FindBy(slug string) *database.Post {
	post := database.Post{}

	result := p.DB.Sql().
		Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Where("LOWER(slug) = ?", slug).
		First(&post)

	if model.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &post
	}

	return nil
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

	if result := p.DB.Sql().Create(&post); model.HasDbIssues(result.Error) {
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

		if result := p.DB.Sql().Create(&trace); model.HasDbIssues(result.Error) {
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

		if result := p.DB.Sql().Create(&trace); model.HasDbIssues(result.Error) {
			return fmt.Errorf("error linking tags [%s:%s]: %s", tag.Name, post.Title, result.Error)
		}
	}

	return nil
}
