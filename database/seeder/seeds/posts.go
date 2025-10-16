package seeds

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/gorm"
)

type PostsSeed struct {
	db *database.Connection
}

func NewPostsSeed(db *database.Connection) *PostsSeed {
	return &PostsSeed{
		db: db,
	}
}

func (s PostsSeed) CreatePosts(attrs database.PostsAttrs, number int) ([]database.Post, error) {
	var posts []database.Post

	for i := 1; i <= number; i++ {
		post := database.Post{
			UUID:          uuid.NewString(),
			AuthorID:      attrs.AuthorID,
			Slug:          attrs.Slug,
			Title:         attrs.Title,
			Excerpt:       "This is an excerpt.",
			Content:       "This is the full content of the post.",
			CoverImageURL: "",
			PublishedAt:   attrs.PublishedAt,
			Categories:    []database.Category{},
			Tags:          []database.Tag{},
			PostViews:     []database.PostView{},
			Comments:      []database.Comment{},
			Likes:         []database.Like{},
		}

		posts = append(posts, post)
	}

	result := s.db.Sql().Create(&posts)

	if gorm.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("issue creating posts: %s", result.Error)
	}

	return posts, nil
}
