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
	t.Run("keeps_https_urls", func(t *testing.T) {
		raw := "https://example.com/path?ok=1#section"
		got := SanitiseURL(raw)
		if got != "https://example.com/path?ok=1" {
			t.Fatalf("expected fragment trimmed https URL, got %q", got)
		}
	})

	t.Run("converts_http_to_https", func(t *testing.T) {
		got := SanitiseURL("http://example.com")
		if got != "https://example.com" {
			t.Fatalf("expected https scheme, got %q", got)
		}
	})

	t.Run("adds_scheme_when_missing", func(t *testing.T) {
		got := SanitiseURL("example.com/page")
		if got != "https://example.com/page" {
			t.Fatalf("expected https scheme added, got %q", got)
		}
	})

	t.Run("allows_localhost", func(t *testing.T) {
		got := SanitiseURL("http://localhost:8080")
		if got != "https://localhost:8080" {
			t.Fatalf("expected localhost to be preserved, got %q", got)
		}
	})

	t.Run("returns_empty_for_invalid_input", func(t *testing.T) {
		for _, input := range []string{"", "   ", "invalid-url", "ftp://example.com"} {
			if got := SanitiseURL(input); got != "" {
				t.Fatalf("expected empty string for %q, got %q", input, got)
			}
		}
	})
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
