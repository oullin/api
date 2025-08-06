package database_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/env"
)

func TestApiKeysWithTestContainer(t *testing.T) {
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

	if err := conn.Sql().AutoMigrate(&database.APIKey{}); err != nil {
		t.Fatalf("migrate err: %v", err)
	}

	repo := repository.ApiKeys{DB: conn}

	created, err := repo.Create(database.APIKeyAttr{
		AccountName: "demo",
		PublicKey:   []byte("pub"),
		SecretKey:   []byte("sec"),
	})

	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if found := repo.FindBy("demo"); found == nil || found.ID != created.ID {
		t.Fatalf("find mismatch")
	}
}
