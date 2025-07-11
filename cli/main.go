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

func init() {
	secrets := boost.Ignite("./../.env", pkg.GetDefaultValidator())

	environment = secrets
	guard = gate.MakeGuard(environment.App.Credentials)
}

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
