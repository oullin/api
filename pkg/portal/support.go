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

// NormalizeOriginWithPath extracts scheme + host + path from a URL string,
// stripping query parameters and fragments. This ensures per-resource signature
// isolation while maintaining consistent matching regardless of query params.
// Returns empty string if the URL is invalid or empty.
func NormalizeOriginWithPath(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Must have a scheme and host
	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	// Return scheme://host/path (without query or fragment)
	path := parsed.Path
	if path == "" {
		path = "/"
	}
	return fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, path)
}

// NormalizeOrigin extracts the base origin (scheme + host) from a URL string,
// stripping path, query parameters, and fragments. Use NormalizeOriginWithPath
// for per-resource signature isolation.
// Returns empty string if the URL is invalid or empty.
func NormalizeOrigin(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Must have a scheme and host
	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}

	// Return just scheme://host
	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
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
