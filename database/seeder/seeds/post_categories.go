package seeds

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/pkg/model"
)

type PostCategoriesSeed struct {
	db *database.Connection
}

func NewPostCategoriesSeed(db *database.Connection) *PostCategoriesSeed {
	return &PostCategoriesSeed{
		db: db,
	}
}

func (s PostCategoriesSeed) Create(category database.Category, post database.Post) error {
	result := s.db.Sql().Create(&database.PostCategory{
		CategoryID: category.ID,
		PostID:     post.ID,
	})

	if model.HasDbIssues(result.Error) {
		return fmt.Errorf("error seeding posts categories: %s", result.Error)
	}

	return nil
}
