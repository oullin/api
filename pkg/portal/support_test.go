package portal

import (
	"net/url"
	"testing"
)

func TestSortedQuery(t *testing.T) {
	u, _ := url.Parse("https://x.test/api?b=2&a=1&a=0")
	got := SortedQuery(u)
	expected := "a=0&a=1&b=2"
	if got != expected {
		t.Fatalf("expected sorted query %q, got %q", expected, got)
	}

	// Empty / nil cases
	if SortedQuery(nil) != "" {
		t.Fatalf("expected empty for nil URL")
	}
	u2, _ := url.Parse("https://x.test/api")
	if SortedQuery(u2) != "" {
		t.Fatalf("expected empty for no query params")
	}
}

func TestBuildCanonical(t *testing.T) {
	u, _ := url.Parse("https://x.test/api/v1/resource?z=9&a=1&a=0")
	bodyHash := "abc123"
	got := BuildCanonical("post", u, "Alice", "pk_123", "1700000000", "nonce-1", bodyHash)
	expected := "POST\n/api/v1/resource\na=0&a=1&z=9\nAlice\npk_123\n1700000000\nnonce-1\nabc123"
	if got != expected {
		t.Fatalf("unexpected canonical string:\nexpected: %q\n     got: %q", expected, got)
	}

	// Default path handling when URL is nil or empty
	got = BuildCanonical("GET", nil, "u", "p", "1", "n", "h")
	if got != "GET\n/\n\nu\np\n1\nn\nh" {
		t.Fatalf("unexpected canonical for nil URL: %q", got)
	}
}
