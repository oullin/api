package portal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func AllowLineBreaks(text string) string {
	replacer := strings.NewReplacer(
		"&lt;br/&gt;", "<br/>",
		"&lt;br /&gt;", "<br/>",
		"&lt;br&gt;", "<br/>",
	)

	return replacer.Replace(text)
}

func FilterNonEmpty(values []string) []string {
	var out []string

	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			out = append(out, strings.TrimSpace(v))
		}
	}

	return out
}

func SanitiseURL(u string) string {
	trimmed := strings.TrimSpace(u)
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)

	if orig, err := url.Parse(trimmed); err == nil {
		switch scheme := strings.ToLower(orig.Scheme); scheme {
		case "":
		// add https later
		case "http", "https":
		// keep going
		default:
			return ""
		}
	}

	candidate := trimmed

	switch {
	case strings.HasPrefix(lower, "https://"):
		// already https
	case strings.HasPrefix(lower, "http://"):
		candidate = "https://" + trimmed[len("http://"):]
	default:
		candidate = "https://" + trimmed
	}

	parsed, err := url.Parse(candidate)
	if err != nil {
		return ""
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return ""
	}

	if hostname != "localhost" && !strings.Contains(hostname, ".") && net.ParseIP(hostname) == nil {
		return ""
	}

	parsed.User = nil
	parsed.Scheme = "https"

	// Remove fragments to keep canonical representation consistent.
	parsed.Fragment = ""

	return template.HTMLEscapeString(parsed.String())
}

func CloseWithLog(c io.Closer) {
	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		slog.Error("failed to close resource", "err", err)
	}
}

func GenerateURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	if v := r.Header.Get("X-Forwarded-Proto"); v != "" {
		scheme = v
	}

	host := r.Host
	if v := r.Header.Get("X-Forwarded-Host"); v != "" {
		host = v
	}

	return scheme + "://" + host + r.URL.RequestURI()
}

func Sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func SortedQuery(u *url.URL) string {
	if u == nil {
		return ""
	}
	q := u.Query()
	if len(q) == 0 {
		return ""
	}
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		vals := q[k]
		sort.Strings(vals)
		for _, v := range vals {
			pairs = append(pairs, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(pairs, "&")
}

func BuildCanonical(method string, u *url.URL, username, public, ts, nonce, bodyHash string) string {
	path := "/"

	if u != nil && u.Path != "" {
		path = u.EscapedPath()
	}

	query := SortedQuery(u)
	parts := []string{
		strings.ToUpper(method),
		path,
		query,
		username,
		public,
		ts,
		nonce,
		bodyHash,
	}

	return strings.Join(parts, "\n")
}

func ParseClientIP(r *http.Request) string {
	// prefer X-Forwarded-For if present

	xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if xff != "" {
		// take first IP
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}

	return strings.TrimSpace(r.RemoteAddr)
}

// headerValue returns the trimmed value for the provided header key, being
// tolerant of non-canonicalized map keys (e.g. raw map literals in tests).
func headerValue(headers http.Header, key string) string {
	if headers == nil {
		return ""
	}

	value := strings.TrimSpace(headers.Get(key))
	if value != "" {
		return value
	}

	canonicalKey := http.CanonicalHeaderKey(key)
	for k, values := range headers {
		if len(values) == 0 {
			continue
		}

		if strings.EqualFold(k, canonicalKey) || strings.EqualFold(k, key) {
			return strings.TrimSpace(values[0])
		}
	}

	return ""
}

// IntendedOriginFromHeader extracts the intended origin value from request headers.
//
// Precedence:
//  1. X-API-Intended-Origin (custom header used for signing)
//  2. Origin (standard browser header)
//  3. Referer (fallback when Origin is absent)
//
// Values are trimmed to avoid mismatches caused by stray whitespace.
func IntendedOriginFromHeader(headers http.Header) string {
	intended := headerValue(headers, IntendedOriginHeader)
	if intended != "" {
		return intended
	}

	origin := headerValue(headers, "Origin")
	referer := headerValue(headers, "Referer")

	if origin == "" {
		return referer
	}

	// Browsers typically send host-only Origin headers. If a Referer is present with the
	// same scheme/host, prefer the referer so signature validation binds to the specific
	// path instead of relaxing to host-level.
	if referer != "" {
		if originURL, err := url.Parse(origin); err == nil {
			if refererURL, err := url.Parse(referer); err == nil {
				if originURL.Scheme == refererURL.Scheme && originURL.Host == refererURL.Host && refererURL.Path != "" {
					return referer
				}
			}
		}
	}

	return origin
}

// ReadWithSizeLimit reads from an io.Reader with a size limit to prevent DoS attacks.
// It returns the read bytes and any error encountered.
// The default size limit is 5MB.
func ReadWithSizeLimit(reader io.Reader, maxSize ...int64) ([]byte, error) {
	if reader == nil {
		return nil, io.ErrUnexpectedEOF
	}

	// Default size limit is 5MB
	const defaultMaxSize int64 = 5 * 1024 * 1024 // 5MB

	limit := defaultMaxSize
	if len(maxSize) > 0 && maxSize[0] > 0 {
		limit = maxSize[0]
	}

	limitedReader := &io.LimitedReader{R: reader, N: limit + 1}
	data, err := io.ReadAll(limitedReader)

	if int64(len(data)) > limit || err != nil {
		return nil, fmt.Errorf("read exceeds size limit: %d, error: %w", limit, err)
	}

	return data, nil
}

// NormalizeOriginWithPath normalizes a URL to include scheme, host, and path,
// but strips query parameters and fragments. This ensures consistent origin
// matching for signature validation while maintaining per-resource isolation.
//
// Normalization follows RFC 3986:
//   - Scheme and host are lowercased
//   - Query parameters and fragments are removed
//   - Trailing slashes on paths are preserved (path semantics matter)
//   - Percent-encoding is preserved as-is
//
// Examples:
//   - https://example.com/api/social?foo=bar#hash → https://example.com/api/social
//   - HTTPS://Example.COM/api/profile → https://example.com/api/profile
//   - /api/social → /api/social (relative URLs are preserved)
func NormalizeOriginWithPath(origin string) string {
	if origin == "" {
		return ""
	}

	parsed, err := url.Parse(origin)
	if err != nil {
		// Invalid URL - return as-is to fail validation downstream
		return origin
	}

	// Normalize scheme to lowercase (RFC 3986 Section 6.2.2.1)
	parsed.Scheme = strings.ToLower(parsed.Scheme)

	// Normalize host to lowercase (RFC 3986 Section 6.2.2.1)
	parsed.Host = strings.ToLower(parsed.Host)

	// Clear query parameters and fragments
	parsed.RawQuery = ""
	parsed.Fragment = ""

	// Note: We preserve trailing slashes because they have semantic meaning
	// in REST APIs (/api/users/ vs /api/users may be different resources)

	return parsed.String()
}
