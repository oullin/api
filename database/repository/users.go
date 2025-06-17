package repository

import (
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg/gorm"
	"strings"
)

type Users struct {
	Connection *database.Connection
	Env        *env.Environment
}

func (r Users) FindBy(username string) *database.User {
	user := &database.User{}

	result := r.Connection.Sql().
		Where("username = ?", username).
		First(&user)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if strings.Trim(user.UUID, " ") != "" {
		return user
	}

	return nil
}
