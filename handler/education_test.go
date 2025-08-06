package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	tests "github.com/oullin/handler/tests"
)

func TestEducationHandlerHandle(t *testing.T) {
	file := tests.WriteJSON(t, tests.TestEnvelope{Version: "v1", Data: []map[string]string{{"uuid": "1"}}})
	defer os.Remove(file)

	h := MakeEducationHandler(file)
	req := httptest.NewRequest("GET", "/education", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var resp tests.TestEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Version != "v1" {
		t.Fatalf("version %s", resp.Version)
	}

	req2 := httptest.NewRequest("GET", "/education", nil)
	req2.Header.Set("If-None-Match", "\"v1\"")
	rec2 := httptest.NewRecorder()
	if err := h.Handle(rec2, req2); err != nil {
		t.Fatalf("err: %v", err)
	}
	if rec2.Code != http.StatusNotModified {
		t.Fatalf("status %d", rec2.Code)
	}

	badF, _ := os.CreateTemp("", "bad.json")
	badF.WriteString("{")
	badF.Close()
	defer os.Remove(badF.Name())

	bad := MakeEducationHandler(badF.Name())
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", "/education", nil)
	if bad.Handle(rec3, req3) == nil {
		t.Fatalf("expected error")
	}
}

func TestEducationHandlerHandle_Payload(t *testing.T) {
	file := tests.WriteJSON(t, tests.TestEnvelope{Version: "v1", Data: []map[string]string{{"uuid": "1"}}})
	defer os.Remove(file)

	h := MakeEducationHandler(file)
	req := httptest.NewRequest("GET", "/education", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}

	var resp tests.TestEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	arr, ok := resp.Data.([]interface{})
	if !ok || len(arr) != 1 {
		t.Fatalf("unexpected data: %+v", resp.Data)
	}
	m, ok := arr[0].(map[string]interface{})
	if !ok || m["uuid"] != "1" {
		t.Fatalf("unexpected payload: %+v", resp.Data)
	}
}
