package main

import (
	_ "github.com/lib/pq"
	"github.com/oullin/boost"
	"github.com/oullin/env"
	"github.com/oullin/pkg"
	"log/slog"
	"net/http"
)

var environment *env.Environment
var validator *pkg.Validator

func init() {
	secrets, validate := boost.Spark("./.env")

	environment = secrets
	validator = validate
}

func main() {
	dbConnection := boost.MakeDbConnection(environment)
	logs := boost.MakeLogs(environment)
	localSentry := boost.MakeSentry(environment)

	defer (*logs).Close()
	defer (*dbConnection).Close()

	mux := http.NewServeMux()

	app := boost.MakeApp(mux, &boost.App{
		Validator:    validator,
		Logs:         logs,
		DbConnection: dbConnection,
		Env:          environment,
		Mux:          mux,
		Sentry:       localSentry,
	})

	app.RegisterUsers()

	(*dbConnection).Ping()
	slog.Info("Starting new server on :" + environment.Network.HttpPort)

	if err := http.ListenAndServe(environment.Network.GetHostURL(), mux); err != nil {
		slog.Error("Error starting server", "error", err)
		panic("Error starting server." + err.Error())
	}
}
