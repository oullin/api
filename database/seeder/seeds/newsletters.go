package seeds

import (
	"fmt"
	"github.com/oullin/database"
	model "github.com/oullin/pkg/model"
)

type NewslettersSeed struct {
	db *database.Connection
}

func MakeNewslettersSeed(db *database.Connection) *NewslettersSeed {
	return &NewslettersSeed{
		db: db,
	}
}

func (s NewslettersSeed) Create(attrs []database.NewsletterAttrs) error {
	var newsletters []database.Newsletter

	for _, attr := range attrs {
		letter := database.Newsletter{
			FirstName:      attr.FirstName,
			LastName:       attr.LastName,
			Email:          attr.Email,
			SubscribedAt:   attr.SubscribedAt,
			UnsubscribedAt: attr.UnsubscribedAt,
		}

		newsletters = append(newsletters, letter)
	}

	result := s.db.Sql().Create(&newsletters)

	if model.HasDbIssues(result.Error) {
		return fmt.Errorf("error seeding newsletters: %s", result.Error)
	}

	return nil
}
