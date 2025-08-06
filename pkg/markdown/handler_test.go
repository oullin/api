package markdown

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParserFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("data"))
	}))
	defer server.Close()

	p := Parser{Url: server.URL}

	content, err := p.Fetch()

	if err != nil || content != "data" {
		t.Fatalf("fetch failed")
	}
}

func TestParserFetchError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	p := Parser{Url: server.URL}

	if _, err := p.Fetch(); err == nil {
		t.Fatalf("expected error")
	}
}
