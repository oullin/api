package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg/gorm"
)

type Posts struct {
	Connection           *database.Connection
	Env                  *env.Environment
	CategoriesRepository Category
}

func (r Posts) Create(attrs database.PostsAttrs) (*database.Post, error) {
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

	// Todo:
	// 1 - Encapsulate all these DB queries in a DB transaction.
	// 2 - Make sure internal queries abort top level DB transactions.
	if result := r.Connection.Sql().Create(&post); gorm.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("issue creating posts: %s", result.Error)
	}

	if _, err := r.CategoriesRepository.CreateOrUpdate(post, attrs); err != nil {
		return &post, fmt.Errorf("issue creating the given post [%s] category: %s", attrs.Slug, err.Error())
	}

	return &post, nil
}
