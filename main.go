package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	slog.Info("starting server", slog.String("address", addr))

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %w", err)
		}

		return nil
	case sig := <-sigCh:
		slog.Info("shutdown signal received", slog.Any("signal", sig))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("shutting down server", slog.String("address", addr))

	if err := server.Shutdown(ctx); err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, http.ErrServerClosed):
			// expected shutdown path
		case errors.Is(err, context.DeadlineExceeded):
			slog.Warn("graceful shutdown timed out, forcing close", slog.String("address", addr))

			if closeErr := server.Close(); closeErr != nil {
				slog.Error("force close server failed", slog.String("address", addr), "error", closeErr)
			}
		default:
			return fmt.Errorf("shutdown server: %w", err)
		}
	}

	if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve: %w", err)
	}

	slog.Info("server stopped", slog.String("address", addr))

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
