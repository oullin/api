package handler

import (
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oullin/pkg/portal"
)

func TestSignaturesHandlerGenerate_ParseError(t *testing.T) {
	h := SignaturesHandler{Validator: portal.GetDefaultValidator()}
	req := httptest.NewRequest("POST", "/signatures", strings.NewReader("{"))
	rec := httptest.NewRecorder()
	if err := h.Generate(rec, req); err == nil || err.Status != nethttp.StatusBadRequest {
		t.Fatalf("expected parse error, got %#v", err)
	}
}
