package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/oullin/database/seeder/importer"
	"github.com/oullin/metal/env"
	"github.com/oullin/metal/kernel"
	"github.com/oullin/pkg/cli"
	"github.com/oullin/pkg/portal"
)

const defaultSQLDumpPath = "./storage/sql/dump.sql"

var (
	environment *env.Environment
	sentryHub   *portal.Sentry
)

func init() {
	secrets := kernel.Ignite("./.env", portal.GetDefaultValidator())

	environment = secrets
	sentryHub = kernel.NewSentry(environment)
}

func main() {
	if err := run(defaultSQLDumpPath, environment, sentryHub); err != nil {
		cli.Errorln(err.Error())
		os.Exit(1)
	}
}

func run(filePath string, environment *env.Environment, sentryHub *portal.Sentry) error {
	if !environment.App.IsLocal() {
		return fmt.Errorf("sql imports can only run in the local environment (current: %s)", environment.App.Type)
	}

	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("importer: SQL file %s does not exist", filePath)
		}
		return fmt.Errorf("importer: unable to access SQL file %s: %w", filePath, err)
	}

	cli.ClearScreen()

	dbConnection := kernel.NewDbConnection(environment)
	logs := kernel.NewLogs(environment)

	defer sentry.Flush(2 * time.Second)
	defer logs.Close()
	defer dbConnection.Close()
	defer kernel.RecoverWithSentry(sentryHub)

	if err := importer.SeedFromFile(dbConnection, environment, filePath); err != nil {
		return err
	}

	cli.Successln("db seeded successfully from SQL file ...")
	return nil
}
