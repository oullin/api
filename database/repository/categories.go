package repository

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/pkg/model"
)

type Categories struct {
	DB *database.Connection
}

func (c Categories) Get() ([]database.Category, error) {
	var categories []database.Category

	err := c.DB.Sql().
		Model(&database.Category{}).
		Where("categories.deleted_at is null").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

func (c Categories) GetAll(paginate pagination.Paginate) (*pagination.Pagination[database.Category], error) {
	var numItems int64
	var categories []database.Category

	query := c.DB.Sql().
		Model(&database.Category{}).
		Joins("JOIN post_categories ON post_categories.category_id = categories.id").
		Joins("JOIN posts ON posts.id = post_categories.post_id").
		Where("categories.deleted_at is null").
		Where("posts.deleted_at is null").
		Where("posts.published_at is not null")

	group := "categories.id, categories.slug"

	if err := pagination.Count[*int64](&numItems, query, c.DB.GetSession(), group); err != nil {
		return nil, err
	}

	offset := (paginate.Page - 1) * paginate.Limit

	err := query.
		Preload("Posts", "posts.deleted_at IS NULL AND posts.published_at IS NOT NULL").
		Offset(offset).
		Limit(paginate.Limit).
		Order("categories.name asc").
		Group(group).
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	paginate.SetNumItems(numItems)
	result := pagination.NewPagination[database.Category](categories, paginate)

	return result, nil
}

func (c Categories) FindBy(slug string) *database.Category {
	category := database.Category{}

	result := c.DB.Sql().
		Where("LOWER(slug) = ?", strings.ToLower(slug)).
		First(&category)

	if model.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &category
	}

	return nil
}

func (c Categories) CreateOrUpdate(post database.Post, attrs database.PostsAttrs) (*[]database.Category, error) {
	var output []database.Category

	for _, seed := range attrs.Categories {
		exists, err := c.ExistOrUpdate(seed)

		if exists {
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("error creating/updating category [%s]: %s", seed.Name, err)
		}

		category := database.Category{
			UUID:        uuid.NewString(),
			Name:        seed.Name,
			Slug:        seed.Slug,
			Description: seed.Description,
		}

		if result := c.DB.Sql().Create(&category); model.HasDbIssues(result.Error) {
			return nil, fmt.Errorf("error creating category [%s]: %s", seed.Name, result.Error)
		}

		trace := database.PostCategory{
			CategoryID: category.ID,
			PostID:     post.ID,
		}

		if result := c.DB.Sql().Create(&trace); model.HasDbIssues(result.Error) {
			return nil, fmt.Errorf("error creating category trace [%s:%s]: %s", category.Name, post.Title, result.Error)
		}

		output = append(output, category)
	}

	return &output, nil
}

func (c Categories) ExistOrUpdate(seed database.CategoriesAttrs) (bool, error) {
	var category *database.Category

	if category = c.FindBy(seed.Slug); category == nil {
		return false, nil
	}

	if strings.Trim(seed.Name, " ") != "" {
		category.Name = seed.Name
	}

	if strings.Trim(seed.Description, " ") != "" {
		category.Description = seed.Description
	}

	if result := c.DB.Sql().Save(&category); model.HasDbIssues(result.Error) {
		return false, fmt.Errorf("error on exist or update category [%s]: %s", category.Name, result.Error)
	}

	return true, nil
}
