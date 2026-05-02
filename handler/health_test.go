package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/oullin/handler/payload"
	"github.com/oullin/pkg/portal"
)

func TestHealthHandler(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("handle err: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var resp payload.KeepAliveResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Message != "ok" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}

	if _, err := time.Parse(portal.DatesLayout, resp.DateTime); err != nil {
		t.Fatalf("invalid datetime: %v", err)
	}
}
