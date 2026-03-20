package seeds

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/oullin/database"
	"github.com/oullin/internal/shared/model"
	"github.com/oullin/internal/shared/portal"
)

type UsersSeed struct {
	db *database.Connection
}

func NewUsersSeed(db *database.Connection) *UsersSeed {
	return &UsersSeed{
		db: db,
	}
}

func (s UsersSeed) Create(attrs database.UsersAttrs) (database.User, error) {
	pass, err := portal.NewPassword("password")
	if err != nil {
		return database.User{}, fmt.Errorf("failed to generate seed password: %w", err)
	}

	fake := database.User{
		UUID:         uuid.NewString(),
		FirstName:    attrs.Name,
		LastName:     "Tester",
		Username:     attrs.Username,
		DisplayName:  fmt.Sprintf("%s User", attrs.Name),
		Email:        fmt.Sprintf("%s@test.com", strings.Trim(attrs.Username, " ")),
		PasswordHash: pass.GetHash(),
		PublicToken:  uuid.NewString(),
		IsAdmin:      attrs.IsAdmin,
		Bio:          "Software engineer with an eye for detail.",
		VerifiedAt:   time.Now(),
	}

	result := s.db.Sql().Create(&fake)

	if model.HasDbIssues(result.Error) {
		return database.User{}, fmt.Errorf("issues creating users: %s", result.Error)
	}

	return fake, nil
}
