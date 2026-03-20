package repository

import (
	"strings"

	"github.com/oullin/database"
	env "github.com/oullin/internal/app/config"
	"github.com/oullin/internal/shared/model"
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
