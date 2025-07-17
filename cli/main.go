package main

import (
	"fmt"
	"github.com/oullin/boost"
	"github.com/oullin/cli/accounts"
	"github.com/oullin/cli/panel"
	"github.com/oullin/cli/posts"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cli"
	"os"
)

var environment *env.Environment
var dbConn *database.Connection

func init() {
	secrets := boost.Ignite("./../.env", pkg.GetDefaultValidator())

	environment = secrets
	dbConn = boost.MakeDbConnection(environment)
}

func main() {
	cli.ClearScreen()

	menu := panel.MakeMenu()

	for {
		err := menu.CaptureInput()

		if err != nil {
			cli.Errorln(err.Error())
			continue
		}

		switch menu.GetChoice() {
		case 1:
			if err = createBlogPost(menu); err != nil {
				cli.Errorln(err.Error())
				continue
			}

			return
		case 2:
			if err = createNewApiAccount(menu); err != nil {
				cli.Errorln(err.Error())
				continue
			}

			return
		case 3:
			if err = showApiAccount(menu); err != nil {
				cli.Errorln(err.Error())
				continue
			}

			return

		case 4:
			signature := auth.CreateSignatureFrom(
				os.Getenv("ENV_LOCAL_TOKEN_ACCOUNT"),
				os.Getenv("ENV_LOCAL_TOKEN_SECRET"),
			)

			cli.Successln("Signature: " + signature)

			return
		case 5:
			if err = generateAppEncryptionKey(); err != nil {
				cli.Errorln(err.Error())
				continue
			}

			return
		case 0:
			cli.Successln("Goodbye!")
			return
		default:
			cli.Errorln("Unknown option. Try again.")
		}

		cli.Blueln("Press Enter to continue...")

		menu.PrintLine()
	}
}

func createBlogPost(menu panel.Menu) error {
	input, err := menu.CapturePostURL()

	if err != nil {
		return err
	}

	httpClient := pkg.MakeDefaultClient(nil)
	handler := posts.MakeHandler(input, httpClient, dbConn)

	if _, err = handler.NotParsed(); err != nil {
		return err
	}

	return nil
}

func createNewApiAccount(menu panel.Menu) error {
	var err error
	var account string
	var handler *accounts.Handler

	if account, err = menu.CaptureAccountName(); err != nil {
		return err
	}

	if handler, err = accounts.MakeHandler(dbConn, environment); err != nil {
		return err
	}

	if err = handler.CreateAccount(account); err != nil {
		return err
	}

	return nil
}

func showApiAccount(menu panel.Menu) error {
	var err error
	var account string
	var handler *accounts.Handler

	if account, err = menu.CaptureAccountName(); err != nil {
		return err
	}

	if handler, err = accounts.MakeHandler(dbConn, environment); err != nil {
		return err
	}

	if handler.ReadAccount(account) != nil {
		return err
	}

	return nil
}

func generateAppEncryptionKey() error {
	var err error
	var key []byte

	if key, err = auth.GenerateAESKey(); err != nil {
		return err
	}

	cli.Successln("\n  The key was generated successfully.")
	cli.Magentaln(fmt.Sprintf("  > key: %x", key))
	fmt.Println(" ")

	return nil
}
