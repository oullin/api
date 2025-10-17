package main

import (
	"errors"
	"fmt"
	"os"
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

func main() {
	defer sentry.Flush(2 * time.Second)

	if err := run(); err != nil {
		sentry.CurrentHub().CaptureException(err)
		cli.Errorln(err.Error())
		sentry.Flush(2 * time.Second)
		os.Exit(1)
	}
}

func run() error {
	cli.ClearScreen()

	validate := portal.GetDefaultValidator()
	environment, err := kernel.Ignite("./.env", validate)
	if err != nil {
		return fmt.Errorf("ignite environment: %w", err)
	}

	hub := kernel.NewSentry(environment)

	defer kernel.RecoverWithSentry(hub)

	dbConn := kernel.NewDbConnection(environment)
	if dbConn == nil {
		return errors.New("database connection is nil")
	}
	defer dbConn.Close()

	menu := panel.NewMenu()

	for {
		if err := menu.CaptureInput(); err != nil {
			cli.Errorln(err.Error())
			continue
		}

		switch menu.GetChoice() {
		case 1:
			if err := createBlogPost(menu, dbConn); err != nil {
				return err
			}
		case 2:
			if err := createNewApiAccount(menu, dbConn, environment); err != nil {
				return err
			}
		case 3:
			if err := showApiAccount(menu, dbConn, environment); err != nil {
				return err
			}
		case 4:
			if err := generateStaticSEO(dbConn, environment); err != nil {
				return err
			}
		case 5:
			if err := generatePostsSEO(dbConn, environment); err != nil {
				return err
			}
		case 6:
			if err := generatePostSEOForSlug(menu, dbConn, environment); err != nil {
				return err
			}
		case 7:
			if err := printTimestamp(); err != nil {
				return err
			}
		case 0:
			cli.Successln("Goodbye!")
			return nil
		default:
			cli.Errorln("Unknown option. Try again.")
		}

		cli.Blueln("Press Enter to continue...")

		menu.PrintLine()
	}
}

func createBlogPost(menu panel.Menu, dbConn *database.Connection) error {
	input, err := menu.CapturePostURL()
	if err != nil {
		return err
	}

	httpClient := portal.NewDefaultClient(nil)
	handler := posts.NewHandler(input, httpClient, dbConn)

	if _, err = handler.NotParsed(); err != nil {
		return err
	}

	return nil
}

func createNewApiAccount(menu panel.Menu, dbConn *database.Connection, environment *env.Environment) error {
	account, err := menu.CaptureAccountName()
	if err != nil {
		return err
	}

	handler, err := accounts.NewHandler(dbConn, environment)
	if err != nil {
		return err
	}

	if err = handler.CreateAccount(account); err != nil {
		return err
	}

	return nil
}

func showApiAccount(menu panel.Menu, dbConn *database.Connection, environment *env.Environment) error {
	account, err := menu.CaptureAccountName()
	if err != nil {
		return err
	}

	handler, err := accounts.NewHandler(dbConn, environment)
	if err != nil {
		return err
	}

	if err = handler.ShowAccount(account); err != nil {
		return err
	}

	return nil
}

func runSEOGeneration(dbConn *database.Connection, environment *env.Environment, genFunc func(*seo.Generator) error) error {
	gen, err := newSEOGenerator(dbConn, environment)
	if err != nil {
		return err
	}

	return genFunc(gen)
}

func generateStaticSEO(dbConn *database.Connection, environment *env.Environment) error {
	return runSEOGeneration(dbConn, environment, (*seo.Generator).GenerateStaticPages)
}

func generatePostsSEO(dbConn *database.Connection, environment *env.Environment) error {
	return runSEOGeneration(dbConn, environment, (*seo.Generator).GeneratePosts)
}

func generatePostSEOForSlug(menu panel.Menu, dbConn *database.Connection, environment *env.Environment) error {
	slug, err := menu.CapturePostSlug()
	if err != nil {
		return err
	}

	return runSEOGeneration(dbConn, environment, func(gen *seo.Generator) error {
		return gen.GeneratePost(slug)
	})
}

func newSEOGenerator(dbConn *database.Connection, environment *env.Environment) (*seo.Generator, error) {
	gen, err := seo.NewGenerator(
		dbConn,
		environment,
		portal.GetDefaultValidator(),
	)
	if err != nil {
		return nil, err
	}

	return gen, nil
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
