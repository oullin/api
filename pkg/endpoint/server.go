package endpoint

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

	"github.com/rs/cors"
)

// RunServer starts the provided HTTP server, listens for shutdown signals, and
// coordinates a graceful shutdown. The addr parameter is used for structured
// logging to identify the server instance.
func RunServer(addr string, server *http.Server) error {
	if server == nil {
		return errors.New("nil http server")
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

// ServerHandlerConfig describes the dependencies required to construct the
// HTTP handler exposed by the API server.
type ServerHandlerConfig struct {
	Mux          http.Handler
	IsProduction bool
	DevHost      string
	Wrap         func(http.Handler) http.Handler
}

// NewServerHandler constructs the HTTP handler using the provided configuration.
// In development environments it applies permissive CORS settings so the
// client app can communicate with the API, and it optionally wraps the handler
// with Sentry instrumentation when supplied.
func NewServerHandler(cfg ServerHandlerConfig) http.Handler {
	if cfg.Mux == nil {
		return http.NotFoundHandler()
	}

	handler := cfg.Mux

	if !cfg.IsProduction {
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

		origins := []string{"http://localhost:5173"}
		if host := cfg.DevHost; host != "" {
			origins = append(origins, host)
		}

		c := cors.New(cors.Options{
			AllowedOrigins:   origins,
			AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
			AllowedHeaders:   headers,
			AllowCredentials: true,
			Debug:            true,
		})

		handler = c.Handler(handler)
	}

	if cfg.Wrap != nil {
		handler = cfg.Wrap(handler)
	}

	return handler
}
