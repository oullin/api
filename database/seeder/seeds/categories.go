package seeds

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/model"
	"strings"
)

type CategoriesSeed struct {
	db *database.Connection
}

func NewCategoriesSeed(db *database.Connection) *CategoriesSeed {
	return &CategoriesSeed{
		db: db,
	}
}

func (s CategoriesSeed) Create(attrs database.CategoriesAttrs) ([]database.Category, error) {
	var categories []database.Category

	seeds := []string{
		"Tech", "AI", "Leadership", "Innovation",
		"Cloud", "Data", "DevOps", "ML", "Startups", "Engineering",
	}

	for index, seed := range seeds {
		categories = append(categories, database.Category{
			UUID:        uuid.NewString(),
			Name:        seed,
			Slug:        strings.ToLower(seed),
			Description: attrs.Description,
			Sort:        index + 1,
		})
	}

	result := s.db.Sql().Create(&categories)

	if model.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("error seeding categories: %s", result.Error)
	}

	return categories, nil
}
