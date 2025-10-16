package seeds

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/pkg/model"
)

type PostViewsSeed struct {
	db *database.Connection
}

func MakePostViewsSeed(db *database.Connection) *PostViewsSeed {
	return &PostViewsSeed{
		db: db,
	}
}

func (s PostViewsSeed) Create(attrs []database.PostViewsAttr) error {
	for _, attr := range attrs {
		result := s.db.Sql().Create(&database.PostView{
			PostID:    attr.Post.ID,
			UserID:    &attr.User.ID,
			IPAddress: attr.IPAddress,
			UserAgent: attr.UserAgent,
		})

		if model.HasDbIssues(result.Error) {
			return fmt.Errorf("issue creating post views for post: %s", result.Error)
		}
	}

	return nil
}
