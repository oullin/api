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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
	defer app.CloseTracer()
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

	var wrap func(http.Handler) http.Handler
	if sentry := app.GetSentry(); sentry != nil && sentry.Handler != nil {
		wrap = sentry.Handler.Handle
	}

	handler := endpoint.NewServerHandler(endpoint.ServerHandlerConfig{
		Mux:          app.GetMux(),
		IsProduction: app.IsProduction(),
		DevHost:      addr,
		Wrap:         wrap,
	})

	// Wrap handler with OpenTelemetry instrumentation if tracing is enabled
	if env.Tracing.Enabled {
		handler = otelhttp.NewHandler(handler, "http.server",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}),
		)
	}

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
