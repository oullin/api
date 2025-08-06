package handler

import (
	"encoding/json"
	baseHttp "net/http"
	"net/http/httptest"
	"os"
	"testing"

	handlertests "github.com/oullin/handler/tests"
	pkgHttp "github.com/oullin/pkg/http"
)

type fileHandler interface {
	Handle(baseHttp.ResponseWriter, *baseHttp.Request) *pkgHttp.ApiError
}

type fileHandlerTestCase struct {
	make     func(string) fileHandler
	endpoint string
	data     interface{}

	assert func(*testing.T, any)
}

func runFileHandlerTest(t *testing.T, tc fileHandlerTestCase) {
	file := handlertests.WriteJSON(t, handlertests.TestEnvelope{
		Version: "v1",
		Data:    tc.data,
	})
	defer os.Remove(file)

	h := tc.make(file)
	req := httptest.NewRequest("GET", tc.endpoint, nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}

	if rec.Code != baseHttp.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var resp handlertests.TestEnvelope

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Version != "v1" {
		t.Fatalf("version %s", resp.Version)
	}

	tc.assert(t, resp.Data)

	req2 := httptest.NewRequest("GET", tc.endpoint, nil)
	req2.Header.Set("If-None-Match", "\"v1\"")
	rec2 := httptest.NewRecorder()

	if err := h.Handle(rec2, req2); err != nil {
		t.Fatalf("err: %v", err)
	}

	if rec2.Code != baseHttp.StatusNotModified {
		t.Fatalf("status %d", rec2.Code)
	}

	badF, _ := os.CreateTemp("", "bad.json")
	badF.WriteString("{")
	badF.Close()
	defer os.Remove(badF.Name())

	bad := tc.make(badF.Name())
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", tc.endpoint, nil)

	if bad.Handle(rec3, req3) == nil {
		t.Fatalf("expected error")
	}
}

func assertArrayUUID1(t *testing.T, data any) {
	arr, ok := data.([]interface{})

	if !ok || len(arr) != 1 {
		t.Fatalf("unexpected data: %+v", data)
	}

	m, ok := arr[0].(map[string]interface{})

	if !ok || m["uuid"] != "1" {
		t.Fatalf("unexpected payload: %+v", data)
	}
}

func assertNicknameNick(t *testing.T, data any) {
	obj, ok := data.(map[string]interface{})

	if !ok || obj["nickname"] != "nick" {
		t.Fatalf("unexpected payload: %+v", data)
	}
}
