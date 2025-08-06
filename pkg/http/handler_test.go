package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeApiHandler(t *testing.T) {
	h := MakeApiHandler(func(w http.ResponseWriter, r *http.Request) *ApiError {

		return &ApiError{Message: "bad", Status: http.StatusBadRequest}
	})

	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest("GET", "/", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
	var resp ErrorResponse

	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Error == "" || resp.Status != http.StatusBadRequest {
		t.Fatalf("invalid response")
	}
}
