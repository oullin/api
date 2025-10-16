package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

type Tags struct {
	DB *database.Connection
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

	if result := t.DB.Sql().Save(&tag); model.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("error creating tag [%s]: %s", slug, result.Error)
	}

	return &tag, nil
}

func (t Tags) FindBy(slug string) *database.Tag {
	tag := database.Tag{}

	result := t.DB.Sql().
		Where("LOWER(slug) = ?", strings.ToLower(slug)).
		First(&tag)

	if model.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &tag
	}

	return nil
}
