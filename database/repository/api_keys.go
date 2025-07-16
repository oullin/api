package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/gorm"
	"strings"
)

type ApiKeys struct {
	DB *database.Connection
}

func (a ApiKeys) Create(attrs database.APIKeyAttr) (*database.APIKey, error) {
	key := database.APIKey{
		UUID:        uuid.NewString(),
		AccountName: attrs.AccountName,
		PublicKey:   attrs.PublicKey,
		SecretKey:   attrs.SecretKey,
	}

	if result := a.DB.Sql().Create(&key); gorm.HasDbIssues(result.Error) {
		return nil, fmt.Errorf(
			"issue creating the given api key pair [%s, %s]: %s",
			attrs.PublicKey,
			attrs.SecretKey,
			result.Error,
		)
	}

	return &key, nil
}

func (a ApiKeys) FindBy(accountName string) *database.APIKey {
	key := database.APIKey{}

	result := a.DB.Sql().
		Where("LOWER(account_name) = ?", strings.ToLower(accountName)).
		First(&key)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if strings.Trim(key.UUID, " ") != "" {
		return &key
	}

	return nil
}
