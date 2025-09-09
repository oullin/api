package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/handler/payload"
)

func TestKeepAliveHandler(t *testing.T) {
	h := MakeKeepAliveHandler()
	req := httptest.NewRequest("GET", "/keep-alive", nil)
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
	if resp.Message != "alive" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}
