package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg/gorm"
	baseGorm "gorm.io/gorm"
)

type Posts struct {
	DB         *database.Connection
	Env        *env.Environment
	categories Categories
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

	err := p.DB.Transaction(func(db *baseGorm.DB) error {
		// --- Post.
		if result := db.Create(&post); gorm.HasDbIssues(result.Error) {
			return fmt.Errorf("issue creating posts: %s", result.Error)
		}

		// --- Categories.
		if _, err := p.categories.CreateOrUpdate(post, attrs); err != nil {
			return fmt.Errorf("issue creating the given post [%s] category: %s", attrs.Slug, err.Error())
		}

		// --- Returning [nil] commits the whole transaction.
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error creating posts[%s]: %s", attrs.Title, err.Error())
	}

	return &post, nil
}
