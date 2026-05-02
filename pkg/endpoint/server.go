package endpoint

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/felixge/httpsnoop"

	"github.com/oullin/pkg/portal"
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
// CORS is handled by Caddy reverse proxy to avoid duplicate headers.
// The handler is optionally wrapped with Sentry instrumentation when supplied.
func NewServerHandler(cfg ServerHandlerConfig) http.Handler {
	if cfg.Mux == nil {
		return http.NotFoundHandler()
	}

	handler := cfg.Mux

	if cfg.Wrap != nil {
		handler = cfg.Wrap(handler)
	}

	return requestLogHandler{next: handler}
}

type requestLogHandler struct {
	next http.Handler
}

func (h requestLogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	metrics := httpsnoop.CaptureMetrics(h.next, w, r)
	status := metrics.Code

	attrs := []any{
		"method", r.Method,
		"path", r.URL.Path,
		"status", status,
		"duration_ms", time.Since(started).Milliseconds(),
		"bytes", metrics.Written,
		"remote_addr", r.RemoteAddr,
		"request_id", r.Header.Get(portal.RequestIDHeader),
		"user_agent", r.UserAgent(),
	}

	if query := safeRequestQuery(r.URL.Query()); query != "" {
		attrs = append(attrs, "query", query)
	}

	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		attrs = append(attrs, "forwarded_for", forwardedFor)
	}

	if status >= http.StatusInternalServerError {
		slog.Error("http request completed", append(attrs, "status_class", strconv.Itoa(status)[0:1]+"xx")...)
		return
	}

	if status >= http.StatusBadRequest {
		slog.Warn("http request completed", append(attrs, "status_class", strconv.Itoa(status)[0:1]+"xx")...)
		return
	}

	slog.Info("http request completed", attrs...)
}

func safeRequestQuery(values url.Values) string {
	safe := url.Values{}

	for _, key := range []string{"limit", "page"} {
		if v, ok := values[key]; ok {
			safe[key] = v
		}
	}

	return safe.Encode()
}
