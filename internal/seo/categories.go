package seo

import (
	"fmt"
	"strings"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
)

type CategoriesSEO struct {
	DB         *database.Connection
	Repository repository.Categories
}

func NewCategories(db *database.Connection) *CategoriesSEO {
	return &CategoriesSEO{
		DB:         db,
		Repository: repository.Categories{DB: db},
	}
}

func (c *CategoriesSEO) Generate() ([]string, error) {
	var err error
	var items []database.Category

	if items, err = c.Repository.Get(); err != nil {
		return nil, fmt.Errorf("could not get categories: %w", err)
	}

	var categories []string
	for item := range items {
		categories = append(categories, strings.ToLower(items[item].Name))
	}

	return categories, nil
}
