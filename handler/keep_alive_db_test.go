package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/internal/testutil/dbtest"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/portal"
)

func TestKeepAliveDBHandler(t *testing.T) {
	db, _ := dbtest.NewTestDB(t)
	e := env.PingEnvironment{Username: "user", Password: "pass"}
	h := NewKeepAliveDBHandler(&e, db)

	t.Run("valid credentials", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/ping-db", nil)
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
		req := httptest.NewRequest("GET", "/ping-db", nil)
		req.SetBasicAuth("bad", "creds")
		rec := httptest.NewRecorder()
		if err := h.Handle(rec, req); err == nil || err.Status != http.StatusUnauthorized {
			t.Fatalf("expected unauthorized, got %#v", err)
		}
	})

	t.Run("db ping failure", func(t *testing.T) {
		var logs bytes.Buffer
		previous := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
		t.Cleanup(func() {
			slog.SetDefault(previous)
		})

		db.Close()
		req := httptest.NewRequest("GET", "/ping-db", nil)
		req.SetBasicAuth("user", "pass")
		rec := httptest.NewRecorder()
		if err := h.Handle(rec, req); err == nil || err.Status != http.StatusInternalServerError {
			t.Fatalf("expected internal error, got %#v", err)
		}

		got := logs.String()
		if count := strings.Count(got, "level=ERROR"); count != 1 {
			t.Fatalf("expected one error log, got %d: %s", count, got)
		}
		if !strings.Contains(got, `msg="database ping failed"`) {
			t.Fatalf("expected structured db ping failure log, got %s", got)
		}
	})
}
