package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/lib/pq"
	"github.com/oullin/metal/kernel"
	"github.com/oullin/pkg/portal"
	"github.com/rs/cors"
)

func main() {
	if err := run(); err != nil {
		sentry.CurrentHub().CaptureException(err)
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	validate := portal.GetDefaultValidator()
	secrets := kernel.Ignite("./.env", validate)

	app, err := kernel.NewApp(secrets, validate)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	defer sentry.Flush(2 * time.Second)
	defer app.CloseDB()
	defer app.CloseLogs()
	defer app.Recover()

	app.Boot()

	if err := app.GetDB().Ping(); err != nil {
		slog.Error("database ping failed", "error", err)
	}

	env := app.GetEnv()
	slog.Info("starting server", slog.String("address", env.Network.GetHostURL()))

	if err := http.ListenAndServe(env.Network.GetHostURL(), serverHandler(app)); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}

func serverHandler(app *kernel.App) http.Handler {
	if app == nil {
		return http.NotFoundHandler()
	}

	mux := app.GetMux()
	if mux == nil {
		return http.NotFoundHandler()
	}

	var handler http.Handler = mux

	if !app.IsProduction() { // Caddy handles CORS.
		env := app.GetEnv()
		localhost := env.Network.GetHostURL()

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
			"X-API-Intended-Origin", // new
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
