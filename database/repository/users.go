package repository

import (
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg/gorm"
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

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if strings.Trim(user.UUID, " ") != "" {
		return &user
	}

	return nil
}
