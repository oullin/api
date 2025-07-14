package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg/gorm"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

type Tags struct {
	DB  *database.Connection
	Env *env.Environment
}

func (t Tags) FindOrCreate(slug string) (*database.Tag, error) {
	if item := t.FindBy(slug); item != nil {
		return item, nil
	}

	caser := cases.Title(language.English, cases.NoLower)

	tag := database.Tag{
		UUID: uuid.NewString(),
		Slug: slug,
		Name: caser.String(strings.ToLower(slug)),
	}

	if result := t.DB.Sql().Save(&tag); gorm.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("error creating tag [%s]: %s", slug, result.Error)
	}

	return &tag, nil
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
