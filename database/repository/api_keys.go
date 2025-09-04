package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/gorm"
)

const MaxSignaturesTries = 5

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

func (a ApiKeys) CreateSignatureFor(key *database.APIKey, seed []byte, expiresAt time.Time) (*database.APIKeySignatures, error) {
	var item *database.APIKeySignatures

	if item = a.FindSignature(key); item != nil {
		return item, nil
	}

	signature := database.APIKeySignatures{
		UUID:      uuid.NewString(),
		APIKeyID:  key.ID,
		Signature: seed,
		Tries:     MaxSignaturesTries,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if result := a.DB.Sql().Create(&signature); gorm.HasDbIssues(result.Error) {
		return nil, fmt.Errorf(
			"issue creating the given api key signature [%s, %s]: ",
			key.AccountName,
			result.Error,
		)
	}

	return &signature, nil
}

func (a ApiKeys) FindBy(accountName string) *database.APIKey {
	key := database.APIKey{}

	result := a.DB.Sql().
		Where("LOWER(account_name) = ?", strings.ToLower(accountName)).
		First(&key)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &key
	}

	return nil
}

func (a ApiKeys) FindSignature(key *database.APIKey) *database.APIKeySignatures {
	var item database.APIKeySignatures

	result := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("api_key_id = ?", key.ID).
		Where("tries <= ?", MaxSignaturesTries).
		Where("expired_at IS NULL OR expired_at > ?", time.Now()).
		First(&item)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &item
	}

	return nil
}
