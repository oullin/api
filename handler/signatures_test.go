package handler

import (
	"fmt"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apih "github.com/oullin/pkg/http"
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

func TestSignaturesHandlerGenerate_UnknownField(t *testing.T) {
	h := SignaturesHandler{Validator: portal.GetDefaultValidator()}
	body := fmt.Sprintf(`{"nonce":"%s","public_key":"%s","username":"%s","timestamp":%d,"extra":"nope"}`,
		strings.Repeat("a", 32),
		strings.Repeat("b", 64),
		"validuser",
		time.Now().Unix(),
	)
	req := httptest.NewRequest("POST", "/signatures", strings.NewReader(body))
	rec := httptest.NewRecorder()
	if err := h.Generate(rec, req); err == nil || err.Status != nethttp.StatusBadRequest {
		t.Fatalf("expected unknown field error, got %#v", err)
	}
}

func TestSignaturesHandlerGenerate_BodyTooLarge(t *testing.T) {
	h := SignaturesHandler{Validator: portal.GetDefaultValidator()}
	large := `{"nonce":"` + strings.Repeat("a", apih.MaxRequestSize+1) + `"}`
	req := httptest.NewRequest("POST", "/signatures", strings.NewReader(large))
	rec := httptest.NewRecorder()
	if err := h.Generate(rec, req); err == nil || err.Status != nethttp.StatusBadRequest {
		t.Fatalf("expected body too large error, got %#v", err)
	}
}
