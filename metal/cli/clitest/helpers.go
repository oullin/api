package clitest

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/oullin/database"
	"github.com/oullin/metal/env"
)

func NewTestConnection(t *testing.T, models ...interface{}) *database.Connection {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
	if err := exec.Command("docker", "ps").Run(); err != nil {
		t.Skip("docker not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Pinning to postgres:18.1-alpine to avoid CVE-2025-12817/12818 and ensure
	// consistent checksum behaviour (initdb enables checksums by default in PG 18).
	pg, err := postgres.Run(ctx,
		"postgres:18.1-alpine",
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

	conn, err := database.NewConnection(e)
	if err != nil {
		t.Fatalf("new connection: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if len(models) > 0 {
		if err := conn.Sql().AutoMigrate(models...); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	return conn
}

func NewTestEnv() *env.Environment {
	return &env.Environment{App: env.AppEnvironment{MasterKey: uuid.NewString()[:32]}}
}
