package clitest

import (
	"context"
	"os/exec"
	"testing"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/env"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func MakeTestConnection(t *testing.T, models ...interface{}) *database.Connection {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}

	ctx := context.Background()

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
	t.Cleanup(func() { pg.Terminate(ctx) })

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

	if len(models) > 0 {
		if err := conn.Sql().AutoMigrate(models...); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	return conn
}

func MakeTestEnv() *env.Environment {
	return &env.Environment{App: env.AppEnvironment{MasterKey: uuid.NewString()[:32]}}
}
