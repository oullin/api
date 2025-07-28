package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository/pagination"
	"github.com/oullin/pkg/gorm"
	"strings"
)

type Categories struct {
	DB *database.Connection
}

func (c Categories) GetAll(paginate pagination.Paginate) (*pagination.Pagination[database.Category], error) {
	var numItems int64
	var categories []database.Category

	query := c.DB.Sql().
		Model(&database.Category{}).
		Where("categories.deleted_at is null").
		Limit(paginate.Limit).
		Order("categories.name asc")

	if err := pagination.Count[*int64](&numItems, query, c.DB.GetSession(), "categories.id"); err != nil {
		return nil, err
	}

	offset := (paginate.Page - 1) * paginate.Limit

	err := query.Preload("Posts").
		Limit(paginate.Limit).
		Offset(offset).
		Distinct().
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	paginate.SetNumItems(numItems)
	result := pagination.MakePagination[database.Category](categories, paginate)

	return result, nil
}

func (c Categories) FindBy(slug string) *database.Category {
	category := database.Category{}

	result := c.DB.Sql().
		Where("LOWER(slug) = ?", strings.ToLower(slug)).
		First(&category)

	if gorm.HasDbIssues(result.Error) {
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

		if result := c.DB.Sql().Create(&category); gorm.HasDbIssues(result.Error) {
			return nil, fmt.Errorf("error creating category [%s]: %s", seed.Name, result.Error)
		}

		trace := database.PostCategory{
			CategoryID: category.ID,
			PostID:     post.ID,
		}

		if result := c.DB.Sql().Create(&trace); gorm.HasDbIssues(result.Error) {
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

	if result := c.DB.Sql().Save(&category); gorm.HasDbIssues(result.Error) {
		return false, fmt.Errorf("error on exist or update category [%s]: %s", category.Name, result.Error)
	}

	return true, nil
}
