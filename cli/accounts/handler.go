package accounts

import (
	"fmt"
	"github.com/oullin/database"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cli"
)

func (h Handler) CreateAccount(accountName string) error {
	token, err := h.TokenHandler.SetupNewAccount(accountName)

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

	cli.Successln("Account created successfully.\n")

	return nil
}

func (h Handler) ReadAccount(accountName string) error {
	item := h.Tokens.FindBy(accountName)

	if item == nil {
		return fmt.Errorf("the given account [%s] was not found", accountName)
	}

	token, err := h.TokenHandler.DecodeTokensFor(
		item.AccountName,
		item.SecretKey,
		item.PublicKey,
	)

	if err != nil {
		return fmt.Errorf("could not decode the given account [%s] keys: %v", item.AccountName, err)
	}

	cli.Successln("\nThe given account has been found successfully!\n")
	cli.Blueln("   > " + fmt.Sprintf("Account name: %s", token.AccountName))
	cli.Blueln("   > " + fmt.Sprintf("Public Key: %s", token.PublicKey))
	cli.Blueln("   > " + fmt.Sprintf("Secret Key: %s", token.SecretKey))
	cli.Blueln("   > " + fmt.Sprintf("API Signature: %s", auth.CreateSignatureFrom(token.AccountName, token.SecretKey)))
	cli.Warningln("----- Encrypted Values -----")
	cli.Magentaln("   > " + fmt.Sprintf("Public Key: %x", token.EncryptedPublicKey))
	cli.Magentaln("   > " + fmt.Sprintf("Secret Key: %x", token.EncryptedSecretKey))
	fmt.Println(" ")

	return nil
}

func (h Handler) CreateSignature(accountName string) error {
	item := h.Tokens.FindBy(accountName)

	if item == nil {
		return fmt.Errorf("the given account [%s] was not found", accountName)
	}

	token, err := h.TokenHandler.DecodeTokensFor(
		item.AccountName,
		item.SecretKey,
		item.PublicKey,
	)

	if err != nil {
		return fmt.Errorf("could not decode the given account [%s] keys: %v", item.AccountName, err)
	}

	signature := auth.CreateSignatureFrom(token.AccountName, token.SecretKey)

	cli.Successln("\nThe given account has been found successfully!\n")
	cli.Blueln("   > " + fmt.Sprintf("Account name: %s", token.AccountName))
	cli.Blueln("   > " + fmt.Sprintf("Public Key: %s", auth.SafeDisplay(token.PublicKey)))
	cli.Blueln("   > " + fmt.Sprintf("Secret Key: %s", auth.SafeDisplay(token.SecretKey)))
	cli.Warningln("----- Encrypted Values -----")
	cli.Magentaln("   > " + fmt.Sprintf("Signature: %s", signature))
	fmt.Println(" ")

	return nil
}
