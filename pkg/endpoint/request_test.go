package endpoint_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/oullin/pkg/endpoint"
)

type sampleReq struct {
	Name string `json:"name"`
}

func TestParseRequestBody(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("{\"name\":\"bob\"}"))
	v, err := endpoint.ParseRequestBody[sampleReq](r)

	if err != nil || v.Name != "bob" {
		t.Fatalf("parse request body failed: %v %#v", err, v)
	}
}

func TestParseRequestBodyEmpty(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	v, err := endpoint.ParseRequestBody[sampleReq](r)

	if err != nil || v.Name != "" {
		t.Fatalf("expected zero value for empty request body")
	}
}

func TestParseRequestBodyInvalid(t *testing.T) {
	r := httptest.NewRequest("POST", "/", strings.NewReader("{"))
	_, err := endpoint.ParseRequestBody[sampleReq](r)

	if err == nil {
		t.Fatalf("expected error for invalid json")
	}
}
