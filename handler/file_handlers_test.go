package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type profileData struct {
	Version string `json:"version"`
	Data    any    `json:"data"`
}

func writeTempJSON(t *testing.T, v any) string {
	f, err := os.CreateTemp("", "data.json")
	if err != nil {
		t.Fatalf("tmp: %v", err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(v); err != nil {
		t.Fatalf("encode: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestProfileHandlerHandle(t *testing.T) {
	file := writeTempJSON(t, profileData{Version: "v1", Data: map[string]string{"nickname": "nick"}})
	defer os.Remove(file)
	h := MakeProfileHandler(file)

	// ok response
	req := httptest.NewRequest("GET", "/profile", nil)
	rec := httptest.NewRecorder()
	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	// cached request
	req2 := httptest.NewRequest("GET", "/profile", nil)
	req2.Header.Set("If-None-Match", "\"v1\"")
	rec2 := httptest.NewRecorder()
	if err := h.Handle(rec2, req2); err != nil {
		t.Fatalf("err: %v", err)
	}
	if rec2.Code != http.StatusNotModified {
		t.Fatalf("status %d", rec2.Code)
	}

	// error on parse
	badF, _ := os.CreateTemp("", "bad.json")
	badF.WriteString("{invalid")
	badF.Close()
	badFile := badF.Name()
	defer os.Remove(badFile)
	bad := MakeProfileHandler(badFile)
	req3 := httptest.NewRequest("GET", "/profile", nil)
	rec3 := httptest.NewRecorder()
	if bad.Handle(rec3, req3) == nil {
		t.Fatalf("expected error")
	}
}

func TestSocialHandlerHandle(t *testing.T) {
	file := writeTempJSON(t, profileData{Version: "v1", Data: []map[string]string{{"uuid": "1"}}})
	defer os.Remove(file)
	h := MakeSocialHandler(file)

	req := httptest.NewRequest("GET", "/social", nil)
	rec := httptest.NewRecorder()
	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}
