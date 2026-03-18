package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/oullin/internal/testutil/apitest"
	"github.com/oullin/pkg/endpoint"
)

type FileHandler interface {
	Handle(http.ResponseWriter, *http.Request) *endpoint.ApiError
}

type FileHandlerTestCase struct {
	Make     func(string) FileHandler
	Endpoint string
	Fixture  string

	Assert func(*testing.T, any)
}

func RunFileHandlerTest(t *testing.T, tc FileHandlerTestCase) {
	f, err := os.Open(tc.Fixture)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	var expected apitest.TestEnvelope

	if err := json.NewDecoder(f).Decode(&expected); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}

	h := tc.Make(tc.Fixture)
	req := httptest.NewRequest("GET", tc.Endpoint, nil)
	rec := httptest.NewRecorder()

	if err := h.Handle(rec, req); err != nil {
		t.Fatalf("err: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var resp apitest.TestEnvelope

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Version != expected.Version {
		t.Fatalf("version %s", resp.Version)
	}

	tc.Assert(t, resp.Data)

	req2 := httptest.NewRequest("GET", tc.Endpoint, nil)
	req2.Header.Set("If-None-Match", rec.Header().Get("ETag"))
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

	bad := tc.Make(badF.Name())
	rec3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("GET", tc.Endpoint, nil)

	if bad.Handle(rec3, req3) == nil {
		t.Fatalf("expected error")
	}
}

func AssertFirstUUID(expected string) func(*testing.T, any) {
	return func(t *testing.T, data any) {
		arr, ok := data.([]interface{})

		if !ok || len(arr) == 0 {
			t.Fatalf("unexpected data: %+v", data)
		}

		m, ok := arr[0].(map[string]interface{})

		if !ok || m["uuid"] != expected {
			t.Fatalf("unexpected payload: %+v", data)
		}
	}
}

func AssertNickname(expected string) func(*testing.T, any) {
	return func(t *testing.T, data any) {
		obj, ok := data.(map[string]interface{})

		if !ok || obj["nickname"] != expected {
			t.Fatalf("unexpected payload: %+v", data)
		}
	}
}
