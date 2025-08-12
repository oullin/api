package handler

import (
	"bytes"
	"context"
	"encoding/json"
	baseHttp "net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/handler/payload"
	"github.com/oullin/metal/env"
	"github.com/oullin/pkg/auth"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupAuth(t *testing.T) AuthHandler {
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
		AccountName: "alice",
		PublicKey:   []byte("pub"),
		SecretKey:   []byte("secret"),
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	jwtHandler, err := auth.MakeJWTHandler(repo, time.Minute)
	if err != nil {
		t.Fatalf("jwt handler: %v", err)
	}

	return MakeAuthHandler(repo, jwtHandler)
}

func TestAuthHandlerToken(t *testing.T) {
	h := setupAuth(t)

	body, _ := json.Marshal(payload.TokenRequest{AccountName: "alice", SecretKey: "secret"})
	req := httptest.NewRequest(baseHttp.MethodPost, "/auth/token", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	if err := h.Token(rec, req); err != nil {
		t.Fatalf("expected nil got %v", err)
	}

	var resp payload.TokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Token == "" {
		t.Fatalf("expected token in response")
	}
}

func TestAuthHandlerTokenUnauthorized(t *testing.T) {
	h := setupAuth(t)

	body, _ := json.Marshal(payload.TokenRequest{AccountName: "alice", SecretKey: "wrong"})
	req := httptest.NewRequest(baseHttp.MethodPost, "/auth/token", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	err := h.Token(rec, req)
	if err == nil || err.Status != baseHttp.StatusUnauthorized {
		t.Fatalf("expected unauthorized got %#v", err)
	}
}
