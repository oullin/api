package main

import (
	_ "github.com/lib/pq"
	"github.com/oullin/boost"
	"github.com/oullin/pkg"
	"log/slog"
	baseHttp "net/http"
)

var app *boost.App

func init() {
	validate := pkg.GetDefaultValidator()

	secrets := boost.Ignite("./.env", validate)

	app = boost.MakeApp(secrets, validate)
}

func main() {
	defer app.CloseDB()
	defer app.CloseLogs()

	app.Boot()

	// --- Testing
	app.GetDB().Ping()
	slog.Info("Starting new server on :" + app.GetEnv().Network.HttpPort)
	// ---

	if err := baseHttp.ListenAndServe(app.GetEnv().Network.GetHostURL(), app.GetMux()); err != nil {
		slog.Error("Error starting server", "error", err)
		panic("Error starting server." + err.Error())
	}
}
