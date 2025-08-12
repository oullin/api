package auth

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func setupRepo(t *testing.T) *repository.ApiKeys {
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

	return repo
}

func TestJWTHandlerGenerateValidate(t *testing.T) {
	repo := setupRepo(t)

	h, err := MakeJWTHandler(repo, time.Minute)
	if err != nil {
		t.Fatalf("make handler err: %v", err)
	}

	token, err := h.Generate("alice")
	if err != nil {
		t.Fatalf("generate token err: %v", err)
	}

	claims, err := h.Validate(token)
	if err != nil {
		t.Fatalf("validate token err: %v", err)
	}

	if claims.AccountName != "alice" {
		t.Fatalf("expected alice got %s", claims.AccountName)
	}
}

func TestJWTHandlerValidateFail(t *testing.T) {
	repo := setupRepo(t)

	h, err := MakeJWTHandler(repo, time.Minute)
	if err != nil {
		t.Fatalf("make handler err: %v", err)
	}

	if _, err := h.Validate("invalid.token"); err == nil {
		t.Fatalf("expected error for invalid token")
	}
}
