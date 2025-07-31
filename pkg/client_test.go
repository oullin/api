package pkg

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientTransportAndGet(t *testing.T) {
	tr := GetDefaultTransport()
	c := MakeDefaultClient(tr)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	defer srv.Close()

	out, err := c.Get(context.Background(), srv.URL)
	if err != nil || out != "hello" {
		t.Fatalf("get failed: %v %s", err, out)
	}
}

func TestClientGetNil(t *testing.T) {
	var c *Client

	_, err := c.Get(context.Background(), "http://example.com")
	if err == nil {
		t.Fatalf("expected error")
	}
}
