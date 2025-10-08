package main

import (
	"errors"
	"flag"
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

var (
	environment *env.Environment
	sentryHub   *portal.Sentry
)

func init() {
	secrets := kernel.Ignite("./.env", portal.GetDefaultValidator())

	environment = secrets
	sentryHub = kernel.MakeSentry(environment)
}

func main() {
	var filePath string
	flag.StringVar(&filePath, "file", "", "name or path to the SQL file located in ./storage/sql to execute")
	flag.Parse()

	if err := run(filePath); err != nil {
		cli.Errorln(err.Error())
		os.Exit(1)
	}
}

func run(filePath string) error {
	if filePath == "" {
		return errors.New("missing required --file flag pointing to a SQL file")
	}

	if !environment.App.IsLocal() {
		return fmt.Errorf("sql imports can only run in the local environment (current: %s)", environment.App.Type)
	}

	cli.ClearScreen()

	dbConnection := kernel.MakeDbConnection(environment)
	logs := kernel.MakeLogs(environment)

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
