package main

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/oullin/database"
	"github.com/oullin/metal/cli/accounts"
	"github.com/oullin/metal/cli/panel"
	"github.com/oullin/metal/cli/posts"
	"github.com/oullin/metal/cli/seo"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/kernel"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/portal"
)

var environment *env.Environment
var dbConn *database.Connection
var sentryHub *portal.Sentry

func init() {
	secrets := kernel.Ignite("./.env", portal.GetDefaultValidator())

	environment = secrets
	dbConn = kernel.MakeDbConnection(environment)
	sentryHub = kernel.MakeSentry(environment)
}

func main() {
	cli.ClearScreen()

	defer sentry.Flush(2 * time.Second)
	defer kernel.RecoverWithSentry(sentryHub)

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
			if err = generateSEO(); err != nil {
				cli.Errorln(err.Error())
				continue
			}

			return
		case 5:
			if err = printTimestamp(); err != nil {
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

	httpClient := portal.MakeDefaultClient(nil)
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

	if err = handler.ShowAccount(account); err != nil {
		return err
	}

	return nil
}

func generateSEO() error {
	gen, err := seo.NewGenerator(
		dbConn,
		environment,
		portal.GetDefaultValidator(),
	)

	if err != nil {
		return err
	}

	if err = gen.Generate(); err != nil {
		return err
	}

	return nil
}

func printTimestamp() error {
	now := time.Now()

	fmt.Println("--- Timestamps ---")

	// 1. Unix Timestamp (seconds since epoch)
	unixTimestampSeconds := now.Unix()
	fmt.Printf("Unix (seconds): %d\n", unixTimestampSeconds)

	// 2. Unix Timestamp (milliseconds since epoch)
	unixTimestampMillis := now.UnixMilli()
	fmt.Printf("Unix (milliseconds): %d\n", unixTimestampMillis)

	// 3. Unix Timestamp (nanoseconds since epoch)
	unixTimestampNanos := now.UnixNano()
	fmt.Printf("Unix (nanoseconds): %d\n", unixTimestampNanos)

	fmt.Println("\n--- Formatted Strings ---")

	// 4. Standard RFC3339 format (e.g., "2025-09-22T14:10:16+08:00")
	rfc3339Timestamp := now.Format(time.RFC3339)
	fmt.Printf("RFC3339: %s\n", rfc3339Timestamp)

	// 5. Custom format (e.g., "YYYY-MM-DD HH:MM:SS")
	// time for layouts: Mon Jan 2 15:04:05 MST 2006
	customTimestamp := now.Format("2006-01-02 15:04:05")
	fmt.Printf("Custom (YYYY-MM-DD HH:MM:SS): %s\n", customTimestamp)

	return nil
}
