package main

import (
	"fmt"
	"log/slog"
	baseHttp "net/http"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	"github.com/oullin/metal/kernel"
	"github.com/oullin/pkg/portal"
	"github.com/rs/cors"
)

var app *kernel.App

func init() {
	validate := portal.GetDefaultValidator()
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

	//if err := baseHttp.ListenAndServe(app.GetEnv().Network.GetHostURL(), app.GetMux()); err != nil {
	//	slog.Error("Error starting server", "error", err)
	//	panic("Error starting server." + err.Error())
	//}
}

func serverHandler() baseHttp.Handler {
	if app.IsProduction() { // Caddy handles CORS.
		return app.GetMux()
	}

	localhost := app.GetEnv().Network.GetHostURL()

	headers := []string{
		"Accept",
		"Authorization",
		"Content-Type",
		"X-CSRF-Token",
		"User-Agent",
		"X-API-Key",
		"X-API-Username",
		"X-API-Signature",
		"X-API-Timestamp",
		"X-API-Nonce",
		"X-Request-ID",
		"If-None-Match",
		"X-API-Intended-Origin", //new
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{localhost, "http://localhost:5173"},
		AllowedMethods:   []string{baseHttp.MethodGet, baseHttp.MethodPost, baseHttp.MethodPut, baseHttp.MethodDelete, baseHttp.MethodOptions},
		AllowedHeaders:   headers,
		AllowCredentials: true,
		Debug:            true,
	})

	return c.Handler(app.GetMux())
}
