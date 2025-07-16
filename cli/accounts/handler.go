package accounts

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cli"
)

func (h Handler) CreateAccount(accountName string) error {
	token, err := auth.SetupNewAccount(accountName, h.TokenLength)

	if err != nil {
		return fmt.Errorf("failed to create account tokens pair: %v", err)
	}

	_, err = h.Tokens.Create(database.APIKeyAttr{
		AccountName: token.AccountName,
		SecretKey:   token.SecretKey,
		PublicKey:   token.PublicKey,
	})

	if err != nil {
		return fmt.Errorf("failed to create account: %v", err)
	}

	cli.Successln("Account created successfully")

	return nil
}
