package main

import (
	"github.com/oullin/boost"
	"github.com/oullin/cli/accounts"
	"github.com/oullin/cli/panel"
	"github.com/oullin/cli/posts"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/cli"
	"time"
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
			input, err := menu.CapturePostURL()

			if err != nil {
				cli.Errorln(err.Error())
				continue
			}

			httpClient := pkg.MakeDefaultClient(nil)
			handler := posts.MakeHandler(input, httpClient, dbConn)

			if _, err := handler.NotParsed(); err != nil {
				cli.Errorln(err.Error())
				continue
			}

			return
		case 2:
			//input, err := menu.CaptureAccountName()

			//if err != nil {
			//	cli.Errorln(err.Error())
			//	continue
			//}
			//
			//handler := accounts.MakeHandler(dbConn)
			//
			//if err := handler.CreateAccount(input); err != nil {
			//	cli.Errorln(err.Error())
			//	continue
			//}

			if err = createNewAccount(menu); err != nil {
				cli.Errorln(err.Error())
				continue
			}

			return

			//generateToken()\\\
		case 3:
			timeParse()
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

func createNewAccount(menu panel.Menu) error {
	var err error
	var account string

	if account, err = menu.CaptureAccountName(); err != nil {
		return err
	}

	handler := accounts.MakeHandler(dbConn)

	if err = handler.CreateAccount(account); err != nil {
		return err
	}

	return nil
}

func timeParse() {
	s := pkg.MakeStringable("2025-04-12")

	if seed, err := s.ToDatetime(); err != nil {
		panic(err)
	} else {
		cli.Magentaln(seed.Format(time.DateTime))
	}
}
