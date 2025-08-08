package portal

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log/slog"
	"net"
	baseHttp "net/http"
	"net/url"
	"sort"
	"strings"
)

func CloseWithLog(c io.Closer) {
	if err := c.Close(); err != nil {
		slog.Error("failed to close resource", "err", err)
	}
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
