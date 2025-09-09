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
	h := MakePingHandler()
	req := httptest.NewRequest("GET", "/ping", nil)
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
}
