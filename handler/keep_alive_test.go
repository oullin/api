package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/portal"
)

func TestKeepAliveHandler(t *testing.T) {
	e := env.PingEnvironment{Username: "user", Password: "pass"}
	h := NewKeepAliveHandler(&e)

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
		var resp payload.KeepAliveResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Message != "pong" {
			t.Fatalf("unexpected message: %s", resp.Message)
		}
		if _, err := time.Parse(portal.DatesLayout, resp.DateTime); err != nil {
			t.Fatalf("invalid datetime: %v", err)
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
