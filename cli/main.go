package main

import (
	"github.com/oullin/boost"
	"github.com/oullin/cli/gate"
	"github.com/oullin/cli/panel"
	"github.com/oullin/cli/posts"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"github.com/oullin/pkg/cli"
	"os"
	"time"
)

var guard gate.Guard
var environment *env.Environment

// init loads environment secrets from a .env file and initializes the global environment and authentication guard.
func init() {
	secrets, _ := boost.Spark("./../.env")

	environment = secrets
	guard = gate.MakeGuard(environment.App.Credentials)
}

// main is the entry point for the CLI application, handling user authentication and presenting a menu-driven interface for further actions.
func main() {
	cli.ClearScreen()

	if err := guard.CaptureInput(); err != nil {
		cli.Errorln(err.Error())
		return
	}

	if guard.Rejects() {
		cli.Errorln("Invalid credentials")
		os.Exit(1)
	}

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
			handler := posts.MakeHandler(input, httpClient, environment)

			if _, err := handler.NotParsed(); err != nil {
				cli.Errorln(err.Error())
				continue
			}

			return
		case 2:
			showTime()
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

func showTime() {
	now := time.Now().Format("2006-01-02 15:04:05")

	cli.Cyanln("\nThe current time is: " + now)
}

func timeParse() {
	s := pkg.MakeStringable("2025-04-12")

	if seed, err := s.ToDatetime(); err != nil {
		panic(err)
	} else {
		cli.Magentaln(seed.Format(time.DateTime))
	}
}
