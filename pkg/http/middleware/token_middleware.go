package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"io"
	"log/slog"
	"net"
	baseHttp "net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cache"
	"github.com/oullin/pkg/http"
)

const tokenHeader = "X-API-Key"
const usernameHeader = "X-API-Username"
const signatureHeader = "X-API-Signature"
const timestampHeader = "X-API-Timestamp"
const nonceHeader = "X-API-Nonce"
const requestIDHeader = "X-Request-ID"

// Context keys for propagating auth info downstream
// Use unexported custom type to avoid collisions

type contextKey string

const (
	authAccountNameKey contextKey = "auth.account_name"
	requestIdKey       contextKey = "request.id"
)

// --- Phase 2 support types

type rateLimiter struct {
	mu       sync.Mutex
	history  map[string][]time.Time // key: ip+"|"+account -> failures timestamps
	window   time.Duration
	maxFails int
}

func newRateLimiter(window time.Duration, maxFails int) *rateLimiter {
	return &rateLimiter{history: make(map[string][]time.Time), window: window, maxFails: maxFails}
}

func (r *rateLimiter) tooMany(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	slice := r.history[key]
	// prune
	pruned := slice[:0]
	for _, t := range slice {
		if now.Sub(t) <= r.window {
			pruned = append(pruned, t)
		}
	}
	r.history[key] = pruned
	return len(pruned) >= r.maxFails
}

func (r *rateLimiter) fail(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	r.history[key] = append(r.history[key], now)
}

type TokenCheckMiddleware struct {
	ApiKeys      *repository.ApiKeys
	TokenHandler *auth.TokenHandler

	// Phase 2 additions
	nonceCache  *cache.TTLCache
	rateLimiter *rateLimiter

	// Configurable parameters
	clockSkew       time.Duration
	nonceTTL        time.Duration
	failWindow      time.Duration
	maxFailPerScope int
}

func MakeTokenMiddleware(tokenHandler *auth.TokenHandler, apiKeys *repository.ApiKeys) TokenCheckMiddleware {
	return TokenCheckMiddleware{
		ApiKeys:         apiKeys,
		TokenHandler:    tokenHandler,
		nonceCache:      cache.NewTTLCache(),
		rateLimiter:     newRateLimiter(1*time.Minute, 10),
		clockSkew:       5 * time.Minute,
		nonceTTL:        5 * time.Minute,
		failWindow:      1 * time.Minute,
		maxFailPerScope: 10,
	}
}

func (t TokenCheckMiddleware) Handle(next http.ApiHandler) http.ApiHandler {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		reqID := strings.TrimSpace(r.Header.Get(requestIDHeader))

		if reqID == "" {
			return t.getInvalidRequestError()
		}

		logger := slog.With("request_id", reqID, "path", r.URL.Path, "method", r.Method)

		accountName := strings.TrimSpace(r.Header.Get(usernameHeader))
		publicToken := strings.TrimSpace(r.Header.Get(tokenHeader))
		signature := strings.TrimSpace(r.Header.Get(signatureHeader))
		ts := strings.TrimSpace(r.Header.Get(timestampHeader))
		nonce := strings.TrimSpace(r.Header.Get(nonceHeader))

		if accountName == "" || publicToken == "" || signature == "" || ts == "" || nonce == "" {
			logger.Warn("missing authentication headers")
			return t.getInvalidRequestError()
		}

		if err := auth.ValidateTokenFormat(publicToken); err != nil {
			logger.Warn("invalid token format")
			return t.getInvalidTokenFormatError()
		}

		// Validate timestamp skew window
		parsed, err := time.ParseDuration(ts + "s")
		if err != nil {
			// try interpreting as epoch seconds (int)
			var epoch int64
			for _, ch := range ts {
				if ch < '0' || ch > '9' {
					logger.Warn("invalid timestamp format")
					return t.getInvalidRequestError()
				}
			}
			// safe to parse as int64 now
			// custom parse to avoid strconv import
			for _, ch := range ts {
				epoch = epoch*10 + int64(ch-'0')
			}
			now := time.Now().Unix()
			if epoch < now-int64(t.clockSkew.Seconds()) || epoch > now+int64(t.clockSkew.Seconds()) {
				logger.Warn("timestamp outside allowed window")
				return t.getUnauthenticatedError()
			}
		} else {
			_ = parsed // ignore if duration parsed (not expected), kept for completeness
		}

		// Read and hash body, then restore it for downstream
		var bodyBytes []byte
		if r.Body != nil {
			b, readErr := io.ReadAll(r.Body)
			if readErr != nil {
				logger.Warn("unable to read body for signing")
				return t.getInvalidRequestError()
			}
			bodyBytes = b
			r.Body = io.NopCloser(bytes.NewReader(b))
		}
		bodyHash := sha256Hex(bodyBytes)

		// Build canonical request string
		canonical := buildCanonical(r.Method, r.URL, accountName, publicToken, ts, nonce, bodyHash)

		clientIP := parseClientIP(r)

		reject := t.shallReject(logger, accountName, publicToken, signature, canonical, nonce, clientIP)
		if reject {
			return t.getUnauthenticatedError()
		}

		// Update the request context
		ctx := context.WithValue(r.Context(), authAccountNameKey, accountName)
		ctx = context.WithValue(r.Context(), requestIdKey, reqID)
		r = r.WithContext(ctx)

		logger.Info("authentication successful")

		return next(w, r)
	}
}

func (t TokenCheckMiddleware) shallReject(logger *slog.Logger, accountName, publicToken, signature, canonical, nonce, clientIP string) bool {
	// Basic rate limiting on failures per IP/account
	limiterKey := clientIP + "|" + strings.ToLower(accountName)
	if t.rateLimiter != nil && t.rateLimiter.tooMany(limiterKey) {
		logger.Warn("too many authentication failures", "ip", clientIP)
		return true
	}

	var item *database.APIKey
	if item = t.ApiKeys.FindBy(accountName); item == nil {
		if t.rateLimiter != nil {
			t.rateLimiter.fail(limiterKey)
		}
		logger.Warn("account not found")
		return true
	}

	token, err := t.TokenHandler.DecodeTokensFor(
		item.AccountName,
		item.SecretKey,
		item.PublicKey,
	)
	if err != nil {
		if t.rateLimiter != nil {
			t.rateLimiter.fail(limiterKey)
		}
		logger.Error("failed to decode account keys", "account", item.AccountName, "error", err)
		return true
	}

	// Constant-time compare of provided public token vs stored one
	provided := []byte(strings.TrimSpace(publicToken))
	expected := []byte(strings.TrimSpace(token.PublicKey))
	if subtle.ConstantTimeCompare(provided, expected) != 1 {
		if t.rateLimiter != nil {
			t.rateLimiter.fail(limiterKey)
		}
		logger.Warn("public token mismatch", "account", item.AccountName)
		return true
	}

	// Nonce replay protection: check if already used for this account
	if t.nonceCache != nil {
		key := item.AccountName + "|" + nonce
		if t.nonceCache.Used(key) {
			if t.rateLimiter != nil {
				t.rateLimiter.fail(limiterKey)
			}
			logger.Warn("replay detected: nonce already used", "account", item.AccountName)
			return true
		}
	}

	// Compute local signature over canonical request and compare in constant time
	localSignature := auth.CreateSignatureFrom(canonical, token.SecretKey)
	if subtle.ConstantTimeCompare([]byte(signature), []byte(localSignature)) != 1 {
		if t.rateLimiter != nil {
			t.rateLimiter.fail(limiterKey)
		}
		logger.Warn("signature mismatch", "account", item.AccountName)
		return true
	}

	// Mark nonce as used within the TTL
	if t.nonceCache != nil {
		key := item.AccountName + "|" + nonce
		t.nonceCache.Mark(key, t.nonceTTL)
	}

	return false
}

// Helpers
func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func sortedQuery(u *url.URL) string {
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

func buildCanonical(method string, u *url.URL, username, public, ts, nonce, bodyHash string) string {
	path := "/"
	if u != nil && u.Path != "" {
		path = u.EscapedPath()
	}
	query := sortedQuery(u)
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

func parseClientIP(r *baseHttp.Request) string {
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

func (t TokenCheckMiddleware) getInvalidRequestError() *http.ApiError {
	return &http.ApiError{
		Message: "Invalid authentication headers",
		Status:  baseHttp.StatusUnauthorized,
	}
}

func (t TokenCheckMiddleware) getInvalidTokenFormatError() *http.ApiError {
	return &http.ApiError{
		Message: "Invalid credentials",
		Status:  baseHttp.StatusUnauthorized,
	}
}

func (t TokenCheckMiddleware) getUnauthenticatedError() *http.ApiError {
	return &http.ApiError{
		Message: "Invalid credentials",
		Status:  baseHttp.StatusUnauthorized,
	}
}
