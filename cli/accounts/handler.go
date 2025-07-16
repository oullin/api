package accounts

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cli"
)

func (h Handler) CreateAccount(accountName string) error {
	tokenHandler, err := auth.MakeTokenHandler(
		[]byte(h.Env.App.MasterKey),
		auth.AccountNameMinLength,
		auth.TokenMinLength,
	)

	if err != nil {
		return fmt.Errorf("error creating the token handler: %v", err)
	}

	token, err := tokenHandler.SetupNewAccount(accountName)

	if err != nil {
		return fmt.Errorf("failed to create the given account [%s] tokens pair: %v", accountName, err)
	}

	_, err = h.Tokens.Create(database.APIKeyAttr{
		AccountName: token.AccountName,
		SecretKey:   token.EncryptedSecretKey,
		PublicKey:   token.EncryptedPublicKey,
	})

	if err != nil {
		return fmt.Errorf("failed to create account [%s]: %v", accountName, err)
	}

	cli.Successln("Account created successfully")

	return nil
}
