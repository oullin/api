package middleware

import (
	"context"
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
	pkgHttp "github.com/oullin/pkg/http"
	"github.com/oullin/pkg/limiter"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// makeRepo creates a temporary postgres repo with a seeded API key
func makeRepo(t *testing.T, account string) (*repository.ApiKeys, *auth.TokenHandler, *auth.Token) {
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("gorm open: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	if err := db.AutoMigrate(&database.APIKey{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	th, err := auth.MakeTokensHandler(generate32(t))
	if err != nil {
		t.Fatalf("MakeTokensHandler: %v", err)
	}
	seed, err := th.SetupNewAccount(account)
	if err != nil {
		t.Fatalf("SetupNewAccount: %v", err)
	}
	if err := db.Create(&database.APIKey{
		UUID:        uuid.NewString(),
		AccountName: seed.AccountName,
		PublicKey:   seed.EncryptedPublicKey,
		SecretKey:   seed.EncryptedSecretKey,
	}).Error; err != nil {
		t.Fatalf("seed api key: %v", err)
	}
	conn := database.NewConnectionFromGorm(db)
	repo := &repository.ApiKeys{DB: conn}
	return repo, th, seed
}

func TestTokenMiddlewareGuardDependencies(t *testing.T) {
	logger := slogNoop()
	tm := TokenCheckMiddleware{}
	if err := tm.GuardDependencies(logger); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized when dependencies missing")
	}
	tm.ApiKeys, tm.TokenHandler, _ = makeRepo(t, "guard1")
	tm.nonceCache = cache.NewTTLCache()
	tm.rateLimiter = limiter.NewMemoryLimiter(time.Minute, 1)
	if err := tm.GuardDependencies(logger); err != nil {
		t.Fatalf("expected no error when dependencies provided, got %#v", err)
	}
}

func TestTokenMiddleware_PublicTokenMismatch(t *testing.T) {
	repo, th, seed := makeRepo(t, "mismatch")
	tm := MakeTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	next := func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil }
	handler := tm.Handle(next)

	req := makeSignedRequest(t, http.MethodGet, "https://api.test.local/v1/x", "", seed.AccountName, "wrong-"+seed.PublicKey, seed.PublicKey, time.Now(), "nonce-mm", "req-mm")
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	rec := httptest.NewRecorder()
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for public token mismatch, got %#v", err)
	}
}

func TestTokenMiddleware_SignatureMismatch(t *testing.T) {
	repo, th, seed := makeRepo(t, "siggy")
	tm := MakeTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	next := func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil }
	handler := tm.Handle(next)

	req := makeSignedRequest(t, http.MethodPost, "https://api.test.local/v1/x", "body", seed.AccountName, seed.PublicKey, seed.PublicKey, time.Now(), "nonce-sig", "req-sig")
	req.Header.Set("X-Forwarded-For", "1.1.1.1")
	req.Header.Set("X-API-Signature", req.Header.Get("X-API-Signature")+"tamper")
	rec := httptest.NewRecorder()
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for signature mismatch, got %#v", err)
	}
}

func TestTokenMiddleware_NonceReplay(t *testing.T) {
	repo, th, seed := makeRepo(t, "replay")
	tm := MakeTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	tm.nonceTTL = time.Minute
	nextCalled := 0
	next := func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
		nextCalled++
		return nil
	}
	handler := tm.Handle(next)

	req := makeSignedRequest(t, http.MethodPost, "https://api.test.local/v1/x", "{}", seed.AccountName, seed.PublicKey, seed.PublicKey, time.Now(), "nonce-rp", "req-rp")
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
	repo, th, seed := makeRepo(t, "ratey")
	tm := MakeTokenMiddleware(th, repo)
	tm.clockSkew = time.Minute
	nextCalled := 0
	next := func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
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
