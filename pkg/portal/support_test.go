package portal

import (
	"errors"
	"io"
	"net/url"
	"strings"
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

func TestSanitiseURL(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "keeps https urls", input: "https://example.com/path?ok=1#section", want: "https://example.com/path?ok=1"},
		{name: "converts http to https", input: "http://example.com", want: "https://example.com"},
		{name: "allows http substring in query", input: "example.com/path?next=http://ok.test", want: "https://example.com/path?next=http://ok.test"},
		{name: "adds scheme when missing", input: "example.com/page", want: "https://example.com/page"},
		{name: "allows localhost", input: "http://localhost:8080", want: "https://localhost:8080"},
		{name: "empty input", input: "", want: ""},
		{name: "whitespace input", input: "   ", want: ""},
		{name: "invalid tld", input: "invalid-url", want: ""},
		{name: "unsupported scheme", input: "ftp://example.com", want: ""},
		{name: "malformed host with space", input: "http://a b.com", want: ""},
		{name: "malformed ipv6 host", input: "http://[::1", want: ""},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := SanitiseURL(tc.input); got != tc.want {
				t.Errorf("SanitiseURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeOriginWithPath(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "keeps path from URL", input: "https://example.com/api/social", want: "https://example.com/api/social"},
		{name: "strips query params", input: "https://example.com/path?foo=bar", want: "https://example.com/path"},
		{name: "strips fragment", input: "https://example.com/page#section", want: "https://example.com/page"},
		{name: "strips query and fragment", input: "https://example.com:8080/api/v1?key=val#top", want: "https://example.com:8080/api/v1"},
		{name: "preserves port with path", input: "https://example.com:3000/api/endpoint", want: "https://example.com:3000/api/endpoint"},
		{name: "handles localhost with path", input: "http://localhost:8080/api/endpoint", want: "http://localhost:8080/api/endpoint"},
		{name: "adds slash for base URL", input: "https://example.com", want: "https://example.com/"},
		{name: "keeps root path", input: "https://example.com/", want: "https://example.com/"},
		{name: "handles nested paths", input: "https://example.com/api/v1/resource", want: "https://example.com/api/v1/resource"},
		{name: "empty input", input: "", want: ""},
		{name: "whitespace input", input: "   ", want: ""},
		{name: "invalid URL", input: "not-a-valid-url", want: ""},
		{name: "missing scheme", input: "example.com/path", want: ""},
		{name: "missing host", input: "https://", want: ""},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeOriginWithPath(tc.input); got != tc.want {
				t.Errorf("NormalizeOriginWithPath(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeOrigin(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "strips path from URL", input: "https://example.com/api/social", want: "https://example.com"},
		{name: "strips query params", input: "https://example.com/path?foo=bar", want: "https://example.com"},
		{name: "strips fragment", input: "https://example.com/page#section", want: "https://example.com"},
		{name: "strips everything except origin", input: "https://example.com:8080/api/v1?key=val#top", want: "https://example.com:8080"},
		{name: "preserves port", input: "https://example.com:3000", want: "https://example.com:3000"},
		{name: "handles localhost", input: "http://localhost:8080/api/endpoint", want: "http://localhost:8080"},
		{name: "keeps base URL unchanged", input: "https://example.com", want: "https://example.com"},
		{name: "empty input", input: "", want: ""},
		{name: "whitespace input", input: "   ", want: ""},
		{name: "invalid URL", input: "not-a-valid-url", want: ""},
		{name: "missing scheme", input: "example.com/path", want: ""},
		{name: "missing host", input: "https://", want: ""},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := NormalizeOrigin(tc.input); got != tc.want {
				t.Errorf("NormalizeOrigin(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
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

// TestReadWithSizeLimit_NilReader tests that ReadWithSizeLimit returns an error when given a nil reader
func TestReadWithSizeLimit_NilReader(t *testing.T) {
	data, err := ReadWithSizeLimit(nil)

	if data != nil {
		t.Errorf("expected nil data for nil reader, got %v", data)
	}

	if err != io.ErrUnexpectedEOF {
		t.Errorf("expected io.ErrUnexpectedEOF for nil reader, got %v", err)
	}
}

// TestReadWithSizeLimit_DefaultLimit tests reading data within and exceeding the default size limit
func TestReadWithSizeLimit_DefaultLimit(t *testing.T) {
	// Test reading data within the default limit
	smallData := strings.Repeat("a", 1024) // 1KB of data
	reader := strings.NewReader(smallData)

	data, err := ReadWithSizeLimit(reader)

	if err != nil {
		t.Errorf("unexpected error for small data: %v", err)
	}

	if string(data) != smallData {
		t.Errorf("data mismatch for small read")
	}

	// We can't easily test the default 5MB limit in a unit test,
	// but we can test the logic by using a smaller custom limit
}

// TestReadWithSizeLimit_CustomLimit tests reading data with a custom size limit
func TestReadWithSizeLimit_CustomLimit(t *testing.T) {
	// Set a small custom limit for testing
	customLimit := int64(100)

	// Test reading data within the custom limit
	smallData := strings.Repeat("a", 50)
	reader := strings.NewReader(smallData)

	data, err := ReadWithSizeLimit(reader, customLimit)

	if err != nil {
		t.Errorf("unexpected error for data within custom limit: %v", err)
	}

	if string(data) != smallData {
		t.Errorf("data mismatch for read within custom limit")
	}

	// Test reading data exceeding the custom limit
	largeData := strings.Repeat("b", 200) // Exceeds our 100 byte limit
	reader = strings.NewReader(largeData)

	data, err = ReadWithSizeLimit(reader, customLimit)

	if err == nil {
		t.Error("expected error for data exceeding custom limit, got nil")
	}

	if data != nil {
		t.Errorf("expected nil data for exceeded limit, got %v", data)
	}
}

// TestReadWithSizeLimit_ErrorPropagation tests that ReadWithSizeLimit properly propagates errors
func TestReadWithSizeLimit_ErrorPropagation(t *testing.T) {
	// Create a reader that returns an error
	expectedErr := errors.New("read error")
	errorReader := &ErrorReader{Err: expectedErr}

	data, err := ReadWithSizeLimit(errorReader)

	if data != nil {
		t.Errorf("expected nil data for error reader, got %v", data)
	}

	if err == nil || !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("expected error containing %q, got %v", expectedErr, err)
	}
}

// ErrorReader is a mock reader that always returns an error
type ErrorReader struct {
	Err error
}

func (r *ErrorReader) Read(p []byte) (n int, err error) {
	return 0, r.Err
}
