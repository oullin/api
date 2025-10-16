package main

import (
	"fmt"
	"log/slog"
	"net/http"
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
	application, err := kernel.NewApp(secrets, validate)

	if err != nil {
		panic(fmt.Sprintf("init: Error creating application: %s", err))
	}

	app = application
}

func main() {
	defer sentry.Flush(2 * time.Second)
	defer app.CloseDB()
	defer app.CloseLogs()
	defer app.Recover()

	app.Boot()

	// --- Testing
	if err := app.GetDB().Ping(); err != nil {
		slog.Error("database ping failed", "error", err)
	}
	slog.Info("Starting new server on :" + app.GetEnv().Network.HttpPort)
	// ---

	if err := http.ListenAndServe(app.GetEnv().Network.GetHostURL(), serverHandler()); err != nil {
		sentry.CurrentHub().CaptureException(err)
		slog.Error("Error starting server", "error", err)
		panic("Error starting server." + err.Error())
	}
}

func serverHandler() http.Handler {
	mux := app.GetMux()
	if mux == nil {
		return http.NotFoundHandler()
	}

	var handler http.Handler = mux

	if !app.IsProduction() { // Caddy handles CORS.
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
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowedHeaders:   headers,
			AllowCredentials: true,
			Debug:            true,
		})

		handler = c.Handler(handler)
	}

	if sentry := app.GetSentry(); sentry != nil && sentry.Handler != nil {
		handler = sentry.Handler.Handle(handler)
	}

	return handler
}
