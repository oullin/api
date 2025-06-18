package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg/gorm"
	"strings"
)

type Category struct {
	Connection *database.Connection
	Env        *env.Environment
}

func (r Category) FindBy(slug string) *database.Category {
	category := &database.Category{}

	result := r.Connection.Sql().
		Where("slug = ?", slug).
		First(&category)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if strings.Trim(category.UUID, " ") != "" {
		return category
	}

	return nil
}

func (r Category) CreateOrUpdate(post database.Post, attrs database.PostsAttrs) (*[]database.Category, error) {
	var output []database.Category

	for _, seed := range attrs.Categories {
		exists, err := r.ExistOrUpdate(seed)

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

		if result := r.Connection.Sql().Create(&category); gorm.HasDbIssues(result.Error) {
			return nil, fmt.Errorf("error creating category [%s]: %s", seed.Name, result.Error)
		}

		trace := database.PostCategory{
			CategoryID: category.ID,
			PostID:     post.ID,
		}

		if result := r.Connection.Sql().Create(&trace); gorm.HasDbIssues(result.Error) {
			return nil, fmt.Errorf("error creating category trace [%s:%s]: %s", category.Name, post.Title, result.Error)
		}

		output = append(output, category)
	}

	return &output, nil
}

func (r Category) ExistOrUpdate(seed database.CategoriesAttrs) (bool, error) {
	var category *database.Category

	if category = r.FindBy(seed.Slug); category == nil {
		return false, nil
	}

	category.Name = seed.Name
	category.Description = seed.Description

	if result := r.Connection.Sql().Save(&category); gorm.HasDbIssues(result.Error) {
		return false, fmt.Errorf("error on exist or update category [%s]: %s", category.Name, result.Error)
	}

	return true, nil
}
