package endpoint_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/portal"
)

func TestNewServerHandlerLogsRequests(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() {
		slog.SetDefault(previous)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler := endpoint.NewServerHandler(endpoint.ServerHandlerConfig{Mux: mux})
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set(portal.RequestIDHeader, "req-health")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	got := logs.String()
	for _, want := range []string{
		"msg=\"http request completed\"",
		"method=GET",
		"path=/health",
		"status=204",
		"request_id=req-health",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected request log to contain %q, got %q", want, got)
		}
	}
}
