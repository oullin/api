package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/gorm"
	baseGorm "gorm.io/gorm"
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

func (a ApiKeys) CreateSignatureFor(key *database.APIKey, seed []byte, expiresAt time.Time) (*database.APIKeySignatures, error) {
	var item *database.APIKeySignatures

	if item = a.FindSignature(key); item != nil {
		return item, nil
	}

	now := time.Now()
	signature := database.APIKeySignatures{
		UUID:      uuid.NewString(),
		APIKeyID:  key.ID,
		Signature: seed,
		Tries:     MaxSignaturesTries,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := a.DB.Transaction(func(tx *baseGorm.DB) error {
		if result := a.DB.Sql().Create(&signature); gorm.HasDbIssues(result.Error) {
			return fmt.Errorf("issue creating the given api key signature [%s, %s]: ", key.AccountName, result.Error)
		}

		if result := a.DisablePreviousSignatures(key, signature.UUID); result != nil {
			return fmt.Errorf("issue creating the given api key signature [%s, %s]: ", key.AccountName, result.Error())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &signature, nil
}

func (a ApiKeys) FindSignature(key *database.APIKey) *database.APIKeySignatures {
	var item database.APIKeySignatures

	result := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("api_key_id = ?", key.ID).
		Where("tries <= ?", MaxSignaturesTries).
		Where("expires_at > ?", time.Now()).
		First(&item)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &item
	}

	return nil
}

func (a ApiKeys) DisablePreviousSignatures(key *database.APIKey, signatureUUID string) error {
	result := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("api_key_id = ?", key.ID).
		Where("uuid != ?", signatureUUID).
		Where("expired_at is null").
		Update("expired_at", time.Now())

	if gorm.HasDbIssues(result.Error) {
		return result.Error
	}

	return nil
}
