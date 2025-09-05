package repository

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository/repoentity"
	"github.com/oullin/pkg/gorm"
	"github.com/oullin/pkg/portal"
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

func (a ApiKeys) CreateSignatureFor(entity repoentity.APIKeyCreateSignatureFor) (*database.APIKeySignatures, error) {
	var item *database.APIKeySignatures

	if item = a.FindActiveSignatureFor(entity.Key, entity.Origin); item != nil {
		item.ExpiresAt = entity.ExpiresAt
		a.DB.Sql().Save(&item)
	}

	now := time.Now()
	signature := database.APIKeySignatures{
		CreatedAt:    now,
		UpdatedAt:    now,
		Signature:    entity.Seed,
		APIKeyID:     entity.Key.ID,
		ExpiresAt:    entity.ExpiresAt,
		UUID:         uuid.NewString(),
		MaxTries:     portal.MaxSignaturesTries,
		Origin:       entity.Origin,
		CurrentTries: 1,
	}

	username := entity.Key.AccountName
	if result := a.DB.Sql().Create(&signature); gorm.HasDbIssues(result.Error) {
		return nil, fmt.Errorf("issue creating the given api keys signature [%s, %s]: ", username, result.Error)
	}

	if result := a.DisablePreviousSignatures(entity.Key, signature.UUID, entity.Origin); result != nil {
		return nil, fmt.Errorf("issue disabling previous api keys signature [%s, %s]: ", username, result.Error())
	}

	return &signature, nil
}

func (a ApiKeys) FindActiveSignatureFor(key *database.APIKey, origin string) *database.APIKeySignatures {
	var item database.APIKeySignatures

	result := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("expired_at IS NULL").
		Where("api_key_id = ?", key.ID).
		Where("origin = ?", origin).
		Where("current_tries <= ? ", portal.MaxSignaturesTries).
		First(&item)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &item
	}

	return nil
}

func (a ApiKeys) FindSignatureFrom(entity repoentity.FindSignatureFrom) *database.APIKeySignatures {
	var item database.APIKeySignatures

	result := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where("api_key_id = ?", entity.Key.ID).
		Where("signature = ?", entity.Signature).
		Where("expires_at >= ? ", entity.ServerTime).
		Where("origin = ?", entity.Origin).
		Where("expired_at IS NULL").
		Where("current_tries <= max_tries").
		First(&item)

	if gorm.HasDbIssues(result.Error) {
		return nil
	}

	if result.RowsAffected > 0 {
		return &item
	}

	return nil
}

func (a ApiKeys) DisablePreviousSignatures(key *database.APIKey, signatureUUID, origin string) error {
	query := a.DB.Sql().
		Model(&database.APIKeySignatures{}).
		Where(
			a.DB.Sql().
				Where("expired_at IS NULL").Or("current_tries > max_tries"),
		).
		Where("api_key_id = ?", key.ID).
		Where(
			a.DB.Sql().
				Where("origin = ?", origin).
				Or("TRIM(origin) = ''"),
		).
		Where("origin = ?", origin).
		Where("uuid NOT IN (?)", []string{signatureUUID}).
		Update("expired_at", time.Now())

	if gorm.HasDbIssues(query.Error) {
		return query.Error
	}

	return nil
}

func (a ApiKeys) IncreaseSignatureTries(signatureUUID string, currentTries int) error {
	if currentTries < portal.MaxSignaturesTries {
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
		Update("current_tries", currentTries)

	if gorm.HasDbIssues(update.Error) {
		return update.Error
	}

	return nil
}
