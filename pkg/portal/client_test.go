package portal_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oullin/pkg/portal"
)

func TestClientTransportAndGet(t *testing.T) {
	tr := portal.GetDefaultTransport()
	c := portal.NewDefaultClient(tr)

	if c.UserAgent != "oullin.io" {
		t.Fatalf("unexpected default user agent: %s", c.UserAgent)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	out, err := c.Get(context.Background(), srv.URL)

	if err != nil || out != "hello" {
		t.Fatalf("get failed: %v %s", err, out)
	}
}

func TestClientGetNil(t *testing.T) {
	var c *portal.Client

	_, err := c.Get(context.Background(), "https://example.com")

	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestClientOnHeadersAndAbort(t *testing.T) {
	c := portal.NewDefaultClient(nil)
	called := false
	c.OnHeaders = func(req *http.Request) {
		req.Header.Set("X-Test", "ok")
		called = true
	}

	c.AbortOnNone2xx = true

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "ok" {
			t.Fatalf("missing header")
		}

		w.Header().Set("Retry-After", "120")
		w.WriteHeader(500)
		_, _ = w.Write([]byte("server error"))
	}))
	defer srv.Close()

	if _, err := c.Get(context.Background(), srv.URL); err == nil {
		t.Fatalf("expected error")
	}

	if !called {
		t.Fatalf("OnHeaders not called")
	}
}

func TestClientGetResponsePreservesMetadataOnAbort(t *testing.T) {
	c := portal.NewDefaultClient(nil)
	c.AbortOnNone2xx = true

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "120")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("slow down"))
	}))
	defer srv.Close()

	resp, err := c.GetResponse(context.Background(), srv.URL)
	if err == nil {
		t.Fatalf("expected error")
	}

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}

	if resp.Header.Get("Retry-After") != "120" {
		t.Fatalf("unexpected retry-after header %q", resp.Header.Get("Retry-After"))
	}

	if resp.Body != "slow down" {
		t.Fatalf("unexpected body %q", resp.Body)
	}
}

func TestClientGetResponseIncludesHeaders(t *testing.T) {
	c := portal.NewDefaultClient(nil)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Link", `<https://example.test?page=2>; rel="last"`)
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	resp, err := c.GetResponse(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("get response failed: %v", err)
	}

	if resp.Body != "hello" {
		t.Fatalf("unexpected body %q", resp.Body)
	}

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("unexpected status code %d", resp.StatusCode)
	}

	if resp.Header.Get("Link") != `<https://example.test?page=2>; rel="last"` {
		t.Fatalf("unexpected link header %q", resp.Header.Get("Link"))
	}
}
