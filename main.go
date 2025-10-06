package main

import (
	"context"
	"fmt"
	"log/slog"
	baseHttp "net/http"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	"github.com/oullin/database/backup"
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
	defer sentry.Flush(2 * time.Second)
	defer app.CloseDB()
	defer app.CloseLogs()
	defer app.Recover()

	app.Boot()

	backupCtx, cancelBackups := context.WithCancel(context.Background())
	defer cancelBackups()

	if scheduler, err := backup.NewScheduler(app.GetEnv()); err != nil {
		slog.Error("failed to create backup scheduler", "error", err)
		panic("backup scheduler initialization failed")
	} else if err := scheduler.Start(backupCtx); err != nil {
		slog.Error("failed to start backup scheduler", "error", err)
		panic("backup scheduler start failed")
	} else {
		slog.Info("database backup scheduler started", "cron", app.GetEnv().Backup.Cron)
		defer scheduler.Stop()
	}

	// --- Testing
	if err := app.GetDB().Ping(); err != nil {
		slog.Error("database ping failed", "error", err)
	}
	slog.Info("Starting new server on :" + app.GetEnv().Network.HttpPort)
	// ---

	if err := baseHttp.ListenAndServe(app.GetEnv().Network.GetHostURL(), serverHandler()); err != nil {
		sentry.CurrentHub().CaptureException(err)
		slog.Error("Error starting server", "error", err)
		panic("Error starting server." + err.Error())
	}
}

func serverHandler() baseHttp.Handler {
	mux := app.GetMux()
	if mux == nil {
		return baseHttp.NotFoundHandler()
	}

	var handler baseHttp.Handler = mux

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
			AllowedMethods:   []string{baseHttp.MethodGet, baseHttp.MethodPost, baseHttp.MethodPut, baseHttp.MethodDelete, baseHttp.MethodOptions},
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
