package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg/gorm"
)

type Posts struct {
	DB         *database.Connection
	Env        *env.Environment
	Categories *Categories
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

	//@todo Add tags tracking

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
