package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"github.com/getsentry/sentry-go"
	"github.com/oullin/metal/kernel"
	"github.com/oullin/pkg"
	"github.com/rs/cors"
	"log/slog"
	baseHttp "net/http"
	"time"
)

var app *kernel.App

func init() {
	validate := pkg.GetDefaultValidator()
	secrets := kernel.Ignite("./.env", validate)
	application, err := kernel.MakeApp(secrets, validate)

	if err != nil {
		panic(fmt.Sprintf("init: Error creating application: %s", err))
	}

	app = application
}

func main() {
	defer app.CloseDB()
	defer app.CloseLogs()
	defer sentry.Flush(2 * time.Second)

	app.Boot()

	// --- Testing
	app.GetDB().Ping()
	slog.Info("Starting new server on :" + app.GetEnv().Network.HttpPort)
	// ---

	if err := baseHttp.ListenAndServe(app.GetEnv().Network.GetHostURL(), serverHandler()); err != nil {
		slog.Error("Error starting server", "error", err)
		panic("Error starting server." + err.Error())
	}
}

func serverHandler() baseHttp.Handler {
	if app.IsProduction() { // CORS is handled by Caddy.
		return app.GetMux()
	}

	localhost := app.GetEnv().Network.GetHostURL()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{localhost, "http://localhost:5173"},
		AllowedMethods:   []string{baseHttp.MethodGet, baseHttp.MethodPost, baseHttp.MethodPut, baseHttp.MethodDelete, baseHttp.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "User-Agent", "X-API-Key", "X-API-Username", "X-API-Signature", "If-None-Match"},
		AllowCredentials: true,
		Debug:            true,
	})

	return c.Handler(app.GetMux())
}
