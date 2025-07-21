package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/gorm"
	"math"
)

type Posts struct {
	DB         *database.Connection
	Categories *Categories
	Tags       *Tags
}

type PostFilters struct {
	UUID           string
	Slug           string
	Title          string // Will perform a case-insensitive partial match
	AuthorUsername string
	CategorySlug   string
	TagSlug        string
	IsPublished    *bool // Pointer to bool to allow three states: true, false, and not-set (nil)
}

func (p Posts) GetPosts(filters *PostFilters, pagination *PaginatedResult[database.Post]) (*PaginatedResult[database.Post], error) {
	var posts []database.Post
	var totalRecords int64

	query := p.DB.Sql().Model(&database.Post{})

	if filters != nil {
		// Filter by direct fields on the 'posts' table
		if filters.UUID != "" {
			query.Where("posts.uuid = ?", filters.UUID)
		}

		if filters.Slug != "" {
			query.Where("posts.slug = ?", filters.Slug)
		}

		if filters.Title != "" {
			// Use ILIKE for case-insensitive search (PostgreSQL specific).
			// For MySQL, use: query.Where("LOWER(posts.title) LIKE LOWER(?)", "%"+filters.Title+"%")
			query.Where("posts.title ILIKE ?", "%"+filters.Title+"%")
		}

		// Filter by relations using JOINs
		if filters.AuthorUsername != "" {
			// GORM's Joins() uses the struct relation name ("Author").
			query.Joins("Author").Where("Author.username = ?", filters.AuthorUsername)
		}

		if filters.CategorySlug != "" {
			// For many-to-many, an explicit join is often clearer and safer.
			query.Joins("JOIN post_categories ON post_categories.post_id = posts.id").
				Joins("JOIN categories ON categories.id = post_categories.category_id").
				Where("categories.slug = ?", filters.CategorySlug)
		}

		if filters.TagSlug != "" {
			query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
				Joins("JOIN tags ON tags.id = post_tags.tag_id").
				Where("tags.slug = ?", filters.TagSlug)
		}
	}

	// Count the total number of records matching the filters
	if err := query.Distinct("posts.id, posts.published_at").Count(&totalRecords).Error; err != nil {
		return nil, err
	}

	// Set default pagination values if none are provided
	if pagination == nil {
		pagination = &PaginatedResult[database.Post]{
			Page:     1,
			PageSize: 10,
		}
	}

	if pagination.Page <= 0 {
		pagination.Page = 1
	}

	if pagination.PageSize <= 0 {
		pagination.PageSize = 10
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pagination.PageSize)))

	var nextPage *int
	if pagination.Page < totalPages {
		p := pagination.Page + 1
		nextPage = &p
	}

	var prevPage *int
	if pagination.Page > 1 && pagination.Page <= totalPages {
		p := pagination.Page - 1
		prevPage = &p
	}

	// Fetch the data for the current page
	offset := (pagination.Page - 1) * pagination.PageSize
	err := query.Preload("Author").
		Preload("Categories").
		Preload("Tags").
		Order("posts.published_at DESC").
		Limit(pagination.PageSize).
		Offset(offset).
		Distinct().
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	// Assemble the final result
	result := &PaginatedResult[database.Post]{
		Data:         posts,
		TotalRecords: totalRecords,
		CurrentPage:  pagination.Page,
		PageSize:     pagination.PageSize,
		TotalPages:   totalPages,
		NextPage:     nextPage,
		PreviousPage: prevPage,
	}

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
