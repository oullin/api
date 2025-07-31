package http

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type sampleReq struct {
	Name string `json:"name"`
}

func TestParseRequestBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("{\"name\":\"bob\"}"))
	v, err := ParseRequestBody[sampleReq](r)
	if err != nil || v.Name != "bob" {
		t.Fatalf("parse failed: %v %#v", err, v)
	}
}

func TestParseRequestBodyEmpty(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	v, err := ParseRequestBody[sampleReq](r)
	if err != nil || v.Name != "" {
		t.Fatalf("expected zero value")
	}
}

func TestParseRequestBodyInvalid(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("{"))
	_, err := ParseRequestBody[sampleReq](r)
	if err == nil {
		t.Fatalf("expected error")
	}
}
