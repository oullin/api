package portal

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net"
	baseHttp "net/http"
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
	u = strings.TrimSpace(u)
	lowerU := strings.ToLower(u)

	if strings.HasPrefix(lowerU, "https://") {
		return template.HTMLEscapeString(u)
	}

	return template.HTMLEscapeString("https://" + u[7:])
}

func CloseWithLog(c io.Closer) {
	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		slog.Error("failed to close resource", "err", err)
	}
}

func GenerateURL(r *baseHttp.Request) string {
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

func ParseClientIP(r *baseHttp.Request) string {
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
