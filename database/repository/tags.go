package repository

import (
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg/gorm"
	"strings"
)

type Tags struct {
	DB  *database.Connection
	Env *env.Environment
}

func (t Tags) FindBy(slug string) *database.Tag {
	tag := database.Tag{}

	result := t.DB.Sql().
		Where("LOWER(slug) = ?", strings.ToLower(slug)).
		First(&tag)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if strings.Trim(tag.UUID, " ") != "" {
		return &tag
	}

	return nil
}
