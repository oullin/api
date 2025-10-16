package repository

import (
	"github.com/oullin/database"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/model"
	"strings"
)

type Users struct {
	DB  *database.Connection
	Env *env.Environment
}

func (u Users) FindBy(username string) *database.User {
	user := database.User{}

	result := u.DB.Sql().
		Where("LOWER(username) = ?", strings.ToLower(username)).
		First(&user)

	if model.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &user
	}

	return nil
}
