package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oullin/handler/payload"
)

func TestPingHandler(t *testing.T) {
	t.Setenv("PING_USERNAME", "user")
	t.Setenv("PING_PASSWORD", "pass")
	h := MakePingHandler()

	t.Run("valid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping", nil)
		req.SetBasicAuth("user", "pass")
		rec := httptest.NewRecorder()
		if err := h.Handle(rec, req); err != nil {
			t.Fatalf("handle err: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d", rec.Code)
		}
		var resp payload.PingResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Message != "pong" {
			t.Fatalf("unexpected message: %s", resp.Message)
		}
		if _, err := time.Parse("2006-01-02", resp.Date); err != nil {
			t.Fatalf("invalid date: %v", err)
		}
		if _, err := time.Parse("15:04:05", resp.Time); err != nil {
			t.Fatalf("invalid time: %v", err)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping", nil)
		req.SetBasicAuth("bad", "creds")
		rec := httptest.NewRecorder()
		if err := h.Handle(rec, req); err == nil || err.Status != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, got %#v", err)
		}
	})
}
