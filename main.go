package main

import (
	_ "github.com/lib/pq"
	"github.com/oullin/boost"
	"github.com/oullin/pkg/http/middleware"
	"log/slog"
	baseHttp "net/http"
)

var app boost.App

func init() {
	secrets, validate := boost.Ignite("./.env")

	app = boost.MakeApp(secrets, validate)
}

func main() {
	defer app.CloseDB()
	defer app.CloseLogs()

	router := boost.Router{
		Env: app.GetEnv(),
		Mux: baseHttp.NewServeMux(),
		Pipeline: middleware.Pipeline{
			Env: app.GetEnv(),
		},
	}

	router.Profile()

	app.GetDB().Ping()
	slog.Info("Starting new server on :" + app.GetEnv().Network.HttpPort)

	if err := baseHttp.ListenAndServe(app.GetEnv().Network.GetHostURL(), router.Mux); err != nil {
		slog.Error("Error starting server", "error", err)
		panic("Error starting server." + err.Error())
	}
}
