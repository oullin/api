package seeds

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/model"
	"strings"
)

type TagsSeed struct {
	db *database.Connection
}

func NewTagsSeed(db *database.Connection) *TagsSeed {
	return &TagsSeed{
		db: db,
	}
}

func (s TagsSeed) Create() ([]database.Tag, error) {
	var tags []database.Tag
	allowed := []string{
		"Tech", "AI", "Leadership", "Ethics",
		"Automation", "Teamwork", "Agile", "OpenAI", "Scaling", "Future",
	}

	for _, name := range allowed {
		tag := database.Tag{
			UUID: uuid.NewString(),
			Name: name,
			Slug: strings.ToLower(name),
		}

		tags = append(tags, tag)
	}

	result := s.db.Sql().Create(&tags)

	if model.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("issues creating tags: %s", result.Error)
	}

	return tags, nil
}
