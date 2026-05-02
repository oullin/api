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

func TestNewServerHandlerPreservesFlusherWhenUnderlyingWriterSupportsIt(t *testing.T) {
	mux := http.NewServeMux()
	var sawFlusher bool

	mux.HandleFunc("GET /stream", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		sawFlusher = ok
		if ok {
			flusher.Flush()
		}
	})

	handler := endpoint.NewServerHandler(endpoint.ServerHandlerConfig{Mux: mux})
	rec := newFlusherResponseWriter()

	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/stream", nil))

	if !sawFlusher {
		t.Fatal("expected downstream handler to receive http.Flusher")
	}

	if !rec.flushed {
		t.Fatal("expected downstream flush to reach underlying writer")
	}
}

func TestNewServerHandlerDoesNotAddFlusherWhenUnderlyingWriterDoesNotSupportIt(t *testing.T) {
	mux := http.NewServeMux()
	var sawFlusher bool

	mux.HandleFunc("GET /plain", func(w http.ResponseWriter, r *http.Request) {
		_, sawFlusher = w.(http.Flusher)
	})

	handler := endpoint.NewServerHandler(endpoint.ServerHandlerConfig{Mux: mux})
	rec := newPlainResponseWriter()

	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/plain", nil))

	if sawFlusher {
		t.Fatal("expected downstream handler not to receive http.Flusher")
	}
}

type flusherResponseWriter struct {
	*plainResponseWriter
	flushed bool
}

func newFlusherResponseWriter() *flusherResponseWriter {
	return &flusherResponseWriter{plainResponseWriter: newPlainResponseWriter()}
}

func (w *flusherResponseWriter) Flush() {
	w.flushed = true
}

type plainResponseWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newPlainResponseWriter() *plainResponseWriter {
	return &plainResponseWriter{header: http.Header{}}
}

func (w *plainResponseWriter) Header() http.Header {
	return w.header
}

func (w *plainResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}

	return w.body.Write(body)
}

func (w *plainResponseWriter) WriteHeader(status int) {
	if w.status != 0 {
		return
	}

	w.status = status
}

var (
	_ http.ResponseWriter = (*plainResponseWriter)(nil)
	_ http.Flusher        = (*flusherResponseWriter)(nil)
)

func TestNewServerHandlerLogsOnlySafeQueryValues(t *testing.T) {
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
	req := httptest.NewRequest("GET", "/health?page=2&limit=10&token=secret&email=a@example.com", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	got := logs.String()
	for _, want := range []string{
		"query=\"limit=10&page=2\"",
		"limit=10",
		"page=2",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected request log to contain %q, got %q", want, got)
		}
	}

	for _, unwanted := range []string{
		"token",
		"secret",
		"email",
		"a@example.com",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("expected request log not to contain %q, got %q", unwanted, got)
		}
	}
}
