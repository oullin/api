package middleware

import (
        "bytes"
        "context"
        "crypto/rand"
        "net/http"
        "net/http/httptest"
        "os/exec"
        "strconv"
        "testing"
        "time"

        "github.com/testcontainers/testcontainers-go"
        "github.com/testcontainers/testcontainers-go/modules/postgres"

        "github.com/oullin/database"
        "github.com/oullin/database/repository"
        "github.com/oullin/metal/env"
        "github.com/oullin/pkg/auth"
        pkgHttp "github.com/oullin/pkg/http"
        "github.com/oullin/pkg/portal"
)

func TestTokenMiddlewareHandle_RequiresRequestID(t *testing.T) {
	tm := MakeTokenMiddleware(nil, nil)

	handler := tm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	// No X-Request-ID present
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected 401 when X-Request-ID is missing, got %#v", err)
	}
}

func TestTokenMiddlewareHandleInvalid(t *testing.T) {
	tm := MakeTokenMiddleware(nil, nil)

	handler := tm.Handle(func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError { return nil })

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-1")
	// Missing other auth headers triggers invalid request
	if err := handler(rec, req); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized for missing auth headers, got %#v", err)
	}
}

func TestValidateAndGetHeaders_MissingAndInvalidFormat(t *testing.T) {
        tm := MakeTokenMiddleware(nil, nil)
        req := httptest.NewRequest("GET", "/", nil)

        if _, apiErr := tm.ValidateAndGetHeaders(req, "req-1"); apiErr == nil || apiErr.Status != http.StatusUnauthorized {
                t.Fatalf("expected error for missing headers")
        }

        // Set minimal headers but invalid token format
        req.Header.Set("X-API-Username", "alice")
        req.Header.Set("X-API-Key", "badtoken")
        req.Header.Set("X-API-Signature", "sig")
        req.Header.Set("X-API-Timestamp", "1700000000")
        req.Header.Set("X-API-Nonce", "n1")
        if _, apiErr := tm.ValidateAndGetHeaders(req, "req-1"); apiErr == nil || apiErr.Status != http.StatusUnauthorized {
                t.Fatalf("expected error for invalid token format")
        }
}

func TestAttachContext(t *testing.T) {
        tm := MakeTokenMiddleware(nil, nil)
        req := httptest.NewRequest("GET", "/", nil)
        headers := AuthTokenHeaders{AccountName: "Alice", RequestID: "RID-123"}
        r := tm.AttachContext(req, headers)
        if r == req {
                t.Fatalf("expected a new request with updated context")
        }
        if r.Context() == nil {
                t.Fatalf("expected non-nil context")
        }
}

// --- Integration test helpers (copied/adjusted from repository_test.go) ---

// setupDB starts a Postgres testcontainer and returns a live DB connection.
func setupDB(t *testing.T) *database.Connection {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
	if err := exec.Command("docker", "ps").Run(); err != nil {
		t.Skip("docker not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	pg, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("secret"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("container run err: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(context.Background()) })

	host, err := pg.Host(ctx)
	if err != nil {
		t.Fatalf("host err: %v", err)
	}
	port, err := pg.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("port err: %v", err)
	}

	e := &env.Environment{
		DB: env.DBEnvironment{
			UserName:     "test",
			UserPassword: "secret",
			DatabaseName: "testdb",
			Port:         port.Int(),
			Host:         host,
			DriverName:   database.DriverName,
			SSLMode:      "disable",
			TimeZone:     "UTC",
		},
	}

	conn, err := database.MakeConnection(e)
	if err != nil {
		t.Fatalf("make connection: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	if err := conn.Sql().AutoMigrate(&database.APIKey{}); err != nil {
		t.Fatalf("migrate err: %v", err)
	}

	return conn
}

// generate32 returns a 32-byte key for TokenHandler.
func generate32(t *testing.T) []byte {
	t.Helper()
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return []byte("0123456789abcdef0123456789abcdef")
	}
	return buf
}

// makeSignedRequest builds a request with required headers and a valid HMAC signature over the canonical string.
//
// signingKey is the token used to create the signature. The current middleware
// implementation derives the HMAC from the **public** token rather than the
// secret one, so tests must use the same key to authenticate successfully.
func makeSignedRequest(t *testing.T, method, rawURL, body, account, public, signingKey string, ts time.Time, nonce, reqID string) *http.Request {
	t.Helper()
	var bodyBuf *bytes.Buffer
	if body != "" {
		bodyBuf = bytes.NewBufferString(body)
	} else {
		bodyBuf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, rawURL, bodyBuf)
	req.Header.Set("X-Request-ID", reqID)
	req.Header.Set("X-API-Username", account)
	req.Header.Set("X-API-Key", public)
	req.Header.Set("X-API-Timestamp", strconv.FormatInt(ts.Unix(), 10))
	req.Header.Set("X-API-Nonce", nonce)

	bodyHash := portal.Sha256Hex([]byte(body))
	canonical := portal.BuildCanonical(method, req.URL, account, public, req.Header.Get("X-API-Timestamp"), nonce, bodyHash)
	sig := auth.CreateSignatureFrom(canonical, signingKey)
	req.Header.Set("X-API-Signature", sig)
	return req
}

func TestTokenMiddleware_DB_Integration(t *testing.T) {
	conn := setupDB(t)

	// Prepare TokenHandler and seed an account with encrypted keys
	th, err := auth.MakeTokensHandler(generate32(t))
	if err != nil {
		t.Fatalf("MakeTokensHandler: %v", err)
	}
	seed, err := th.SetupNewAccount("acme-user")
	if err != nil {
		t.Fatalf("SetupNewAccount: %v", err)
	}

	repo := &repository.ApiKeys{DB: conn}
	if _, err := repo.Create(database.APIKeyAttr{
		AccountName: seed.AccountName,
		PublicKey:   seed.EncryptedPublicKey,
		SecretKey:   seed.EncryptedSecretKey,
	}); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}

	// Build middleware
	tm := MakeTokenMiddleware(th, repo)
	// make it tolerant and fast for test
	tm.clockSkew = 2 * time.Minute
	tm.nonceTTL = 1 * time.Minute

	nextCalled := false
	next := func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
		nextCalled = true
		return nil
	}
	handler := tm.Handle(next)

	// Positive case
	now := time.Now()
	req := makeSignedRequest(t,
		http.MethodPost,
		"https://api.test.local/v1/posts?z=9&a=1",
		"{\"title\":\"ok\"}",
		seed.AccountName,
		seed.PublicKey,
		seed.PublicKey,
		now,
		"nonce-1",
		"req-001",
	)
	rec := httptest.NewRecorder()
	if err := handler(rec, req); err != nil {
		t.Fatalf("expected success, got error: %#v", err)
	}
	if !nextCalled {
		t.Fatalf("expected next to be called on success")
	}

	// Negative case: unknown account
	nextCalled = false
	reqUnknown := makeSignedRequest(t,
		http.MethodGet,
		"https://api.test.local/v1/ping",
		"",
		"no-such-user",
		seed.PublicKey,
		seed.PublicKey,
		now,
		"nonce-2",
		"req-002",
	)
	rec = httptest.NewRecorder()
	if err := handler(rec, reqUnknown); err == nil || err.Status != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unknown account, got %#v", err)
	}
	if nextCalled {
		t.Fatalf("next should not be called on auth failure")
	}
}

// New happy path only test
func TestTokenMiddleware_DB_Integration_HappyPath(t *testing.T) {
	conn := setupDB(t)

	// Prepare TokenHandler and seed an account with encrypted keys
	th, err := auth.MakeTokensHandler(generate32(t))
	if err != nil {
		t.Fatalf("MakeTokensHandler: %v", err)
	}
	seed, err := th.SetupNewAccount("acme-user-happy")
	if err != nil {
		t.Fatalf("SetupNewAccount: %v", err)
	}

	repo := &repository.ApiKeys{DB: conn}
	if _, err := repo.Create(database.APIKeyAttr{
		AccountName: seed.AccountName,
		PublicKey:   seed.EncryptedPublicKey,
		SecretKey:   seed.EncryptedSecretKey,
	}); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}

	// Build middleware
	tm := MakeTokenMiddleware(th, repo)
	// Relax window for test
	tm.clockSkew = 2 * time.Minute
	tm.nonceTTL = 1 * time.Minute

	nextCalled := false
	next := func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
		nextCalled = true
		return nil
	}
	handler := tm.Handle(next)

	req := makeSignedRequest(t,
		http.MethodPost,
		"https://api.test.local/v1/resource?b=2&a=1",
		"{\"x\":123}",
		seed.AccountName,
		seed.PublicKey,
		seed.PublicKey,
		time.Now(),
		"n-happy-1",
		"rid-happy-1",
	)
	rec := httptest.NewRecorder()
	if err := handler(rec, req); err != nil {
		t.Fatalf("happy path failed: %#v", err)
	}
	if !nextCalled {
		t.Fatalf("next was not called on happy path")
	}
}

// TestMakeTokenMiddleware_DefaultDisallowFuture verifies that disallowFuture is true by default
func TestMakeTokenMiddleware_DefaultDisallowFuture(t *testing.T) {
	// Create middleware with default settings
	tm := MakeTokenMiddleware(nil, nil)

	// Verify disallowFuture is true by default
	if !tm.disallowFuture {
		t.Fatalf("expected disallowFuture to be true by default, got false")
	}
}

// TestTokenMiddleware_RejectsFutureTimestamps verifies that future timestamps are rejected
func TestTokenMiddleware_RejectsFutureTimestamps(t *testing.T) {
	conn := setupDB(t)

	// Prepare TokenHandler and seed an account with encrypted keys
	th, err := auth.MakeTokensHandler(generate32(t))
	if err != nil {
		t.Fatalf("MakeTokensHandler: %v", err)
	}
	seed, err := th.SetupNewAccount("acme-user-future")
	if err != nil {
		t.Fatalf("SetupNewAccount: %v", err)
	}

	repo := &repository.ApiKeys{DB: conn}
	if _, err := repo.Create(database.APIKeyAttr{
		AccountName: seed.AccountName,
		PublicKey:   seed.EncryptedPublicKey,
		SecretKey:   seed.EncryptedSecretKey,
	}); err != nil {
		t.Fatalf("repo.Create: %v", err)
	}

	// Build middleware with default settings (disallowFuture = true)
	tm := MakeTokenMiddleware(th, repo)

	nextCalled := false
	next := func(w http.ResponseWriter, r *http.Request) *pkgHttp.ApiError {
		nextCalled = true
		return nil
	}
	handler := tm.Handle(next)

	// Create a request with a future timestamp (30 seconds in the future)
	futureTime := time.Now().Add(30 * time.Second)
	req := makeSignedRequest(t,
		http.MethodGet,
		"https://api.test.local/v1/test",
		"",
		seed.AccountName,
		seed.PublicKey,
		seed.PublicKey,
		futureTime,
		"n-future-1",
		"rid-future-1",
	)
	rec := httptest.NewRecorder()

	// The request should be rejected with a 401 Unauthorized
	apiErr := handler(rec, req)
	if apiErr == nil || apiErr.Status != http.StatusUnauthorized {
		t.Fatalf("expected 401 for future timestamp, got %#v", apiErr)
	}

	// Next handler should not be called
	if nextCalled {
		t.Fatalf("next should not be called when future timestamp is rejected")
	}
}
