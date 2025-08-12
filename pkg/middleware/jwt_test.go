package middleware

import (
	"context"
	baseHttp "net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/auth"
	"github.com/oullin/pkg/http"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupJWTHandler(t *testing.T) auth.JWTHandler {
	t.Helper()

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
	t.Cleanup(func() { pg.Terminate(context.Background()) })

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
	t.Cleanup(func() { conn.Close() })

	if err := conn.Sql().AutoMigrate(&database.APIKey{}); err != nil {
		t.Fatalf("migrate err: %v", err)
	}

	repo := &repository.ApiKeys{DB: conn}
	if _, err := repo.Create(database.APIKeyAttr{
		AccountName: "bob",
		PublicKey:   []byte("pub"),
		SecretKey:   []byte("mysecretjwtkey12345"),
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	h, err := auth.MakeJWTHandler(repo, time.Minute)
	if err != nil {
		t.Fatalf("make handler err: %v", err)
	}

	return h
}

func TestJWTMiddlewareHandle(t *testing.T) {
	handler := setupJWTHandler(t)
	m := JWTMiddleware{Handler: handler}

	next := func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError {
		claims, ok := GetJWTClaims(r.Context())
		if !ok {
			t.Fatalf("claims missing from context")
		}
		if claims.AccountName != "bob" {
			t.Fatalf("expected bob got %s", claims.AccountName)
		}
		return nil
	}

	token, err := handler.Generate("bob")
	if err != nil {
		t.Fatalf("generate token err: %v", err)
	}

	req := httptest.NewRequest(baseHttp.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	if err := m.Handle(next)(rr, req); err != nil {
		t.Fatalf("expected nil got %v", err)
	}
}

func TestJWTMiddlewareUnauthorized(t *testing.T) {
	handler := setupJWTHandler(t)
	m := JWTMiddleware{Handler: handler}

	req := httptest.NewRequest(baseHttp.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	errResp := m.Handle(func(w baseHttp.ResponseWriter, r *baseHttp.Request) *http.ApiError { return nil })(rr, req)
	if errResp == nil {
		t.Fatalf("expected error for missing token")
	}
	if errResp.Status != baseHttp.StatusUnauthorized {
		t.Fatalf("expected unauthorized got %d", errResp.Status)
	}
}
