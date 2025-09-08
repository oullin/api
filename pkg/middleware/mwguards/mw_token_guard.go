package mwguards

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/auth"
)

type MWTokenGuard struct {
	Error          error
	ApiKey         *database.APIKey
	TokenHandler   *auth.TokenHandler
	KeysRepository *repository.ApiKeys
}

type MWTokenGuardData struct {
	Username  string
	PublicKey string
}

func NewMWTokenGuard(apiKeys *repository.ApiKeys, TokenHandler *auth.TokenHandler) MWTokenGuard {
	return MWTokenGuard{
		KeysRepository: apiKeys,
		TokenHandler:   TokenHandler,
	}
}

func (g *MWTokenGuard) Rejects(data MWTokenGuardData) bool {
	if g.HasInvalidDependencies() {
		g.Error = fmt.Errorf("invalid mw-token guard dependencies")

		return true
	}

	if err := g.AccountNotFound(data.Username); err != nil {
		g.Error = err

		return true
	}

	if g.HasInvalidFormat(data.PublicKey) {
		return true
	}

	return false
}

func (g *MWTokenGuard) HasInvalidDependencies() bool {
	return g == nil || g.KeysRepository == nil || g.TokenHandler == nil
}

func (g *MWTokenGuard) AccountNotFound(username string) error {
	var item *database.APIKey

	if item = g.KeysRepository.FindBy(username); item == nil {
		return fmt.Errorf("account [%s] not found", username)
	}

	g.ApiKey = item

	return nil
}

func (g *MWTokenGuard) HasInvalidFormat(publicKey string) bool {
	token, err := g.TokenHandler.DecodeTokensFor(
		g.ApiKey.AccountName,
		g.ApiKey.SecretKey,
		g.ApiKey.PublicKey,
	)

	if err != nil {
		g.Error = fmt.Errorf("unable to decode the given account [%s] keys", g.ApiKey.AccountName)

		return true
	}

	pBytes := []byte(publicKey)
	eBytes := []byte(token.PublicKey)
	hP := sha256.Sum256(pBytes)
	hE := sha256.Sum256(eBytes)

	if subtle.ConstantTimeCompare(hP[:], hE[:]) != 1 {
		g.Error = fmt.Errorf("invalid provided public token: %s", auth.SafeDisplay(publicKey))

		return true
	}

	return false
}

func (g *MWTokenGuard) GetError() error {
	return g.Error
}

func (receiver MWTokenGuardData) ToMap() map[string]any {
	return map[string]any{
		"username":   receiver.Username,
		"public_key": receiver.PublicKey,
	}
}
