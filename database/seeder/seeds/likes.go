package seeds

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/model"
)

type LikesSeed struct {
	db *database.Connection
}

func NewLikesSeed(db *database.Connection) *LikesSeed {
	return &LikesSeed{
		db: db,
	}
}

func (s LikesSeed) Create(attrs ...database.LikesAttrs) ([]database.Like, error) {
	var likes []database.Like

	for _, attr := range attrs {
		likes = append(likes, database.Like{
			UUID:   uuid.NewString(),
			PostID: attr.PostID,
			UserID: attr.UserID,
		})
	}

	result := s.db.Sql().Create(&likes)

	if model.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("error seeding likes: %s", result.Error)
	}

	return likes, nil
}
