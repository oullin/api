package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestProfileHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/profile", map[string]string{"nickname": "nick"}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeProfileHandler(p)
		return h
	})
}

func TestProfileHandlerHandle_Payload(t *testing.T) {
	file := writeJSON(t, testEnvelope{Version: "v1", Data: map[string]string{"nickname": "nick"}})
	defer os.Remove(file)

	h := MakeProfileHandler(file)
	req := httptest.NewRequest("GET", "/profile", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}

	var resp testEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	obj, ok := resp.Data.(map[string]interface{})
	if !ok || obj["nickname"] != "nick" {
		t.Fatalf("unexpected payload: %+v", resp.Data)
	}
}
