package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/pkg/gorm"
	"github.com/oullin/pkg/portal"
	baseGorm "gorm.io/gorm"
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

	if result.RowsAffected > 0 {
		return &key
	}

	return nil
}

func (a ApiKeys) CreateSignatureFor(key *database.APIKey, seed []byte, expiresAt time.Time) (*database.APIKeySignatures, error) {
	var item *database.APIKeySignatures

	if item = a.FindActiveSignatureFor(key); item != nil {
		return item, nil
	}

	now := time.Now()
	signature := database.APIKeySignatures{
		CreatedAt: now,
		UpdatedAt: now,
		Signature: seed,
		APIKeyID:  key.ID,
		ExpiresAt: expiresAt,
		UUID:      uuid.NewString(),
		Tries:     portal.MaxSignaturesTries,
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

func (a ApiKeys) FindActiveSignatureFor(key *database.APIKey) *database.APIKeySignatures {
	var item database.APIKeySignatures

	result := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("expired_at IS NULL").
		Where("api_key_id = ?", key.ID).
		Where("tries <=", portal.MaxSignaturesTries).
		First(&item)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &item
	}

	return nil
}

func (a ApiKeys) FindSignatureFrom(key *database.APIKey, signature []byte) *database.APIKeySignatures {
	var item database.APIKeySignatures

	result := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("api_key_id = ?", key.ID).
		Where("signature = ?", signature).
		Where("expired_at IS NULL").
		Where("tries <=", portal.MaxSignaturesTries).
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
	query := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("expired_at IS NULL").
		Where("api_key_id = ?", key.ID).
		Where("uuid NOT IN (?)", []string{signatureUUID}).
		Update("expired_at", time.Now())

	if gorm.HasDbIssues(query.Error) {
		return query.Error
	}

	return nil
}

func (a ApiKeys) IncreaseSignatureTries(signatureUUID string, tries int) error {
	if tries < portal.MaxSignaturesTries {
		return nil
	}

	var item database.APIKeySignatures

	query := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("uuid = ?", signatureUUID).
		First(&item)

	if gorm.HasDbIssues(query.Error) {
		return query.Error
	}

	update := a.DB.Sql().
		Model(&item).
		Update("tries", tries)

	if gorm.HasDbIssues(update.Error) {
		return update.Error
	}

	return nil
}
