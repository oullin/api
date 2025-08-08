package main

import (
	"fmt"

	"github.com/oullin/database"
	"github.com/oullin/metal/cli/accounts"
	"github.com/oullin/metal/cli/panel"
	"github.com/oullin/metal/cli/posts"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/kernel"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/portal"
)

var environment *env.Environment
var dbConn *database.Connection

func init() {
	secrets := kernel.Ignite("./.env", portal.GetDefaultValidator())

	environment = secrets
	dbConn = kernel.MakeDbConnection(environment)
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
			if err = generateApiAccountsHTTPSignature(menu); err != nil {
				cli.Errorln(err.Error())
				continue
			}

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

	if err = handler.ReadAccount(account); err != nil {
		return err
	}

	return nil
}

func generateApiAccountsHTTPSignature(menu panel.Menu) error {
	var err error
	var account string
	var handler *accounts.Handler

	if account, err = menu.CaptureAccountName(); err != nil {
		return err
	}

	if handler, err = accounts.MakeHandler(dbConn, environment); err != nil {
		return err
	}

	if err = handler.CreateSignature(account); err != nil {
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

	decoded := fmt.Sprintf("%x", key)

	cli.Successln("\n  The key was generated successfully.")
	cli.Magentaln(fmt.Sprintf("  > Full key: %s", decoded))
	cli.Cyanln(fmt.Sprintf("  > First half : %s", decoded[:32]))
	cli.Cyanln(fmt.Sprintf("  > Second half: %s", decoded[32:]))
	fmt.Println(" ")

	return nil
}
