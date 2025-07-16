package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/gorm"
)

type ApiKeys struct {
	DB *database.Connection
}

func (p ApiKeys) Create(attrs database.APIKeyAttr) (*database.APIKey, error) {
	key := database.APIKey{
		UUID:        uuid.NewString(),
		AccountName: attrs.AccountName,
		PublicKey:   attrs.PublicKey,
		SecretKey:   attrs.SecretKey,
	}

	if result := p.DB.Sql().Create(&key); gorm.HasDbIssues(result.Error) {
		return nil, fmt.Errorf(
			"issue creating the given api key pair [%s, %s]: %s",
			attrs.PublicKey,
			attrs.SecretKey,
			result.Error,
		)
	}

	return &key, nil
}
