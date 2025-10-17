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
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"
	"github.com/rs/cors"
)

func main() {
	defer sentry.Flush(2 * time.Second)

	if err := run(); err != nil {
		sentry.CurrentHub().CaptureException(err)
		if !sentry.Flush(2 * time.Second) {
			slog.Warn("sentry flush timed out after capture")
		}
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	validate := portal.GetDefaultValidator()
	secrets, err := kernel.Ignite("./.env", validate)
	if err != nil {
		return fmt.Errorf("ignite environment: %w", err)
	}

	app, err := kernel.NewApp(secrets, validate)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	defer app.CloseDB()
	defer app.CloseLogs()
	defer app.Recover()

	app.Boot()

	if err := app.GetDB().Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	env := app.GetEnv()
	if env == nil {
		return errors.New("application environment is nil")
	}
	addr := env.Network.GetHostURL()
	handler := serverHandler(app)

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := endpoint.RunServer(addr, server); err != nil {
		return fmt.Errorf("serve http: %w", err)
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

	handler := http.Handler(mux)

	if !app.IsProduction() { // Caddy handles CORS.
		env := app.GetEnv()
		if env == nil {
			return handler
		}
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
			"X-API-Intended-Origin",
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
