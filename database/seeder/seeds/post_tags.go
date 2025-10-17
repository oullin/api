package seeds

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/pkg/model"
)

type PostTagsSeed struct {
	db *database.Connection
}

func NewPostTagsSeed(db *database.Connection) *PostTagsSeed {
	return &PostTagsSeed{
		db: db,
	}
}

func (s PostTagsSeed) Create(tag database.Tag, post database.Post) error {
	result := s.db.Sql().Create(&database.PostTag{
		PostID: post.ID,
		TagID:  tag.ID,
	})

	if model.HasDbIssues(result.Error) {
		return fmt.Errorf("error seeding tags: %s", result.Error)
	}

	return nil
}
