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

func TestNormalizeOriginWithPath(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strips query parameters",
			input: "https://example.com/api/social?foo=bar&baz=qux",
			want:  "https://example.com/api/social",
		},
		{
			name:  "strips fragment",
			input: "https://example.com/api/profile#section",
			want:  "https://example.com/api/profile",
		},
		{
			name:  "strips both query and fragment",
			input: "https://example.com/api/social?foo=bar#section",
			want:  "https://example.com/api/social",
		},
		{
			name:  "preserves path without query or fragment",
			input: "https://example.com/api/social",
			want:  "https://example.com/api/social",
		},
		{
			name:  "handles root path",
			input: "https://example.com/",
			want:  "https://example.com/",
		},
		{
			name:  "handles relative URL with query",
			input: "/api/social?foo=bar",
			want:  "/api/social",
		},
		{
			name:  "handles relative URL without query",
			input: "/api/social",
			want:  "/api/social",
		},
		{
			name:  "handles empty string",
			input: "",
			want:  "",
		},
		{
			name:  "preserves different paths",
			input: "https://example.com/api/profile?param=value",
			want:  "https://example.com/api/profile",
		},
		{
			name:  "handles localhost",
			input: "http://localhost:8080/api/test?debug=true",
			want:  "http://localhost:8080/api/test",
		},
		{
			name:  "normalizes uppercase scheme to lowercase (RFC 3986)",
			input: "HTTPS://example.com/api/social",
			want:  "https://example.com/api/social",
		},
		{
			name:  "normalizes uppercase host to lowercase (RFC 3986)",
			input: "https://EXAMPLE.COM/api/social",
			want:  "https://example.com/api/social",
		},
		{
			name:  "normalizes mixed case scheme and host",
			input: "HTTPS://Example.COM/api/Social?foo=bar",
			want:  "https://example.com/api/Social",
		},
		{
			name:  "preserves path case sensitivity",
			input: "https://example.com/API/Social",
			want:  "https://example.com/API/Social",
		},
		{
			name:  "preserves trailing slash",
			input: "https://example.com/api/social/",
			want:  "https://example.com/api/social/",
		},
		{
			name:  "normalizes with port number",
			input: "https://Example.COM:8080/api/social?test=1",
			want:  "https://example.com:8080/api/social",
		},
		{
			name:  "handles percent-encoded characters",
			input: "https://example.com/api/social%20media?foo=bar",
			want:  "https://example.com/api/social%20media",
		},
		{
			name:  "handles space-containing string (percent-encodes)",
			input: "not a valid URL at all",
			want:  "not%20a%20valid%20URL%20at%20all",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeOriginWithPath(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeOriginWithPath(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
