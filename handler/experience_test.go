package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	pkghttp "github.com/oullin/pkg/http"
)

func TestExperienceHandlerHandle(t *testing.T) {
	runFileHandlerTest(t, "/experience", []map[string]string{{"uuid": "1"}}, func(p string) interface {
		Handle(http.ResponseWriter, *http.Request) *pkghttp.ApiError
	} {
		h := MakeExperienceHandler(p)
		return h
	})
}

func TestExperienceHandlerHandle_Payload(t *testing.T) {
	file := writeJSON(t, testEnvelope{Version: "v1", Data: []map[string]string{{"uuid": "1"}}})
	defer os.Remove(file)

	h := MakeExperienceHandler(file)
	req := httptest.NewRequest("GET", "/experience", nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}

	var resp testEnvelope
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
