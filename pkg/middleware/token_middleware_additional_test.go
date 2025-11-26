package middleware

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/cache"
	"github.com/oullin/pkg/endpoint"
	"github.com/oullin/pkg/limiter"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// makeRepo creates a temporary postgres repo with a seeded API key
func makeRepo(t *testing.T, account string) (*repository.ApiKeys, *auth.TokenHandler, *auth.Token, *database.APIKey) {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pgC, err := postgrescontainer.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgrescontainer.WithDatabase("testdb"),
		postgrescontainer.WithUsername("test"),
		postgrescontainer.WithPassword("test"),
	)
	if err != nil {
		t.Skipf("run postgres container: %v", err)
	}
	t.Cleanup(func() {
		cctx, ccancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer ccancel()
		_ = pgC.Terminate(cctx)
	})
	dsn, err := pgC.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Skipf("connection string: %v", err)
	}
	var db *gorm.DB
	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		t.Skipf("gorm open: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	if err := db.AutoMigrate(&database.APIKey{}, &database.APIKeySignatures{}); err != nil {
		t.Skipf("migrate: %v", err)
	}
	th, err := auth.NewTokensHandler(generate32(t))
	if err != nil {
		t.Fatalf("NewTokensHandler: %v", err)
	}
	seed, err := th.SetupNewAccount(account)
	if err != nil {
		t.Fatalf("SetupNewAccount: %v", err)
	}
	key := database.APIKey{
		UUID:        uuid.NewString(),
		AccountName: seed.AccountName,
		PublicKey:   seed.EncryptedPublicKey,
		SecretKey:   seed.EncryptedSecretKey,
	}
	if err := db.Create(&key).Error; err != nil {
		t.Skipf("seed api key: %v", err)
	}
	conn := database.NewConnectionFromGorm(db)
	repo := &repository.ApiKeys{DB: conn}
	return repo, th, seed, &key
}

func TestTokenMiddlewareGuardDependencies(t *testing.T) {
	tm := TokenCheckMiddleware{}
	if err := tm.GuardDependencies(); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized when dependencies missing")
	}
	repo, th, _, _ := makeRepo(t, "guard1")
	tm.ApiKeys, tm.TokenHandler = repo, th
	tm.nonceCache = cache.NewTTLCache()
	tm.rateLimiter = limiter.NewMemoryLimiter(time.Minute, 1)
	if err := tm.GuardDependencies(); err != nil {
		t.Fatalf("expected no error when dependencies provided, got %#v", err)
	}
}

func TestTokenMiddleware_PublicTokenMismatch(t *testing.T) {
	repo, th, seed, _ := makeRepo(t, "mismatch")
	tm := NewTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	next := func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError { return nil }
	handler := tm.Handle(next)

	req := makeSignedRequest(t, http.MethodGet, "https://api.test.local/v1/x", "", seed.AccountName, "wrong-"+seed.PublicKey, seed.PublicKey, time.Now(), "nonce-mm", "req-mm")
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	rec := httptest.NewRecorder()
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for public token mismatch, got %#v", err)
	}
}

func TestTokenMiddleware_SignatureMismatch(t *testing.T) {
	repo, th, seed, key := makeRepo(t, "siggy")
	tm := NewTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	next := func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError { return nil }
	handler := tm.Handle(next)

	req := makeSignedRequest(t, http.MethodPost, "https://api.test.local/v1/x", "body", seed.AccountName, seed.PublicKey, seed.PublicKey, time.Now(), "nonce-sig", "req-sig")
	seedSignature(t, repo, key, req, time.Now().Add(time.Hour))
	req.Header.Set("X-Forwarded-For", "1.1.1.1")

	// mutate signature while keeping valid hex encoding
	sigHex := req.Header.Get("X-API-Signature")
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	sigBytes[0] ^= 0xFF
	req.Header.Set("X-API-Signature", hex.EncodeToString(sigBytes))
	rec := httptest.NewRecorder()
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for signature mismatch, got %#v", err)
	}
}

func TestTokenMiddleware_NonceReplay(t *testing.T) {
	repo, th, seed, key := makeRepo(t, "replay")
	tm := NewTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	tm.nonceTTL = time.Minute
	nextCalled := 0
	next := func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
		nextCalled++
		return nil
	}
	handler := tm.Handle(next)

	req := makeSignedRequest(t, http.MethodPost, "https://api.test.local/v1/x", "{}", seed.AccountName, seed.PublicKey, seed.PublicKey, time.Now(), "nonce-rp", "req-rp")
	seedSignature(t, repo, key, req, time.Now().Add(time.Hour))
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	rec := httptest.NewRecorder()
	if err := handler(rec, req); err != nil {
		t.Fatalf("first call failed: %#v", err)
	}
	if nextCalled != 1 {
		t.Fatalf("expected next called once on first request")
	}
	rec = httptest.NewRecorder()
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized on nonce replay, got %#v", err)
	}
}

func TestTokenMiddleware_RateLimiter(t *testing.T) {
	repo, th, seed, key := makeRepo(t, "ratey")
	tm := NewTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	nextCalled := 0
	next := func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
		nextCalled++
		return nil
	}
	handler := tm.Handle(next)

	// Pre-warm limiter by sending invalid signature requests up to the limit
	for i := 0; i < tm.maxFailPerScope; i++ {
		req := makeSignedRequest(
			t, http.MethodGet, "https://api.test.local/v1/rl", "",
			seed.AccountName, seed.PublicKey, "wrong-secret", time.Now(),
			fmt.Sprintf("nonce-rl-%d", i), fmt.Sprintf("req-rl-%d", i),
		)
		req.Header.Set("X-Forwarded-For", "9.9.9.9")
		rec := httptest.NewRecorder()
		_ = handler(rec, req) // ignore errors while warming
	}

	// Next request with valid signature should be rate limited
	req := makeSignedRequest(
		t, http.MethodGet, "https://api.test.local/v1/rl", "",
		seed.AccountName, seed.PublicKey, seed.PublicKey, time.Now(),
		"nonce-rl-final", "req-rl-final",
	)
	seedSignature(t, repo, key, req, time.Now().Add(time.Hour))
	req.Header.Set("X-Forwarded-For", "9.9.9.9")
	rec := httptest.NewRecorder()
	err := handler(rec, req)
	if err == nil || err.Status != http.StatusTooManyRequests {
		t.Fatalf("expected rate limited error, got %#v", err)
	}

	if nextCalled != 0 {
		t.Fatalf("expected next not to be invoked when rate limited, got %d calls", nextCalled)
	}
}

func TestTokenMiddleware_CustomClockValidatesSignature(t *testing.T) {
	repo, th, seed, key := makeRepo(t, "clock")
	tm := NewTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	past := time.Now().Add(-10 * time.Minute)
	tm.now = func() time.Time { return past }

	nextCalled := false
	handler := tm.Handle(func(w http.ResponseWriter, r *http.Request) *endpoint.ApiError {
		nextCalled = true
		return nil
	})

	req := makeSignedRequest(t, http.MethodGet, "https://api.test.local/v1/clock", "", seed.AccountName, seed.PublicKey, seed.PublicKey, past, "nonce-clock", "req-clock")
	seedSignature(t, repo, key, req, past.Add(5*time.Minute))
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	rec := httptest.NewRecorder()
	if err := handler(rec, req); err != nil {
		t.Fatalf("expected success with injected clock, got %#v", err)
	}
	if !nextCalled {
		t.Fatalf("expected next to be called")
	}
}
