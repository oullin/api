package sqlseed_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/oullin/database"
	"github.com/oullin/database/seeder/sqlseed"
	"github.com/oullin/metal/env"
)

func TestSeedFromFileExecutesStatements(t *testing.T) {
	conn, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	dir := t.TempDir()
	path := filepath.Join(dir, "seed.sql")
	sql := "CREATE TABLE widgets (id SERIAL PRIMARY KEY, name TEXT NOT NULL);\nINSERT INTO widgets (name) VALUES ('alpha'), ('beta');"

	if err := os.WriteFile(path, []byte(sql), 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	if err := sqlseed.SeedFromFile(conn, path); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	var count int64
	if err := conn.Sql().Table("widgets").Count(&count).Error; err != nil {
		t.Fatalf("count widgets: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 widgets, got %d", count)
	}
}

func TestSeedFromFileRejectsNonSQLFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.txt")

	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	err := sqlseed.SeedFromFile(nil, path)
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Fatalf("expected extension error, got %v", err)
	}
}

func TestSeedFromFileFailsWhenFileMissing(t *testing.T) {
	err := sqlseed.SeedFromFile(nil, filepath.Join(t.TempDir(), "missing.sql"))
	if err == nil || !strings.Contains(err.Error(), "read file") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestSeedFromFileFailsWhenFileEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.sql")

	if err := os.WriteFile(path, []byte("   \n\t"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := sqlseed.SeedFromFile(nil, path)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected empty file error, got %v", err)
	}
}

func TestSeedFromFileRequiresConnection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "seed.sql")

	if err := os.WriteFile(path, []byte("SELECT 1;"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := sqlseed.SeedFromFile(nil, path)
	if err == nil || !strings.Contains(err.Error(), "connection") {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func TestSeedFromFileRollsBackOnFailure(t *testing.T) {
	conn, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	dir := t.TempDir()
	path := filepath.Join(dir, "seed.sql")
	sql := "CREATE TABLE gadgets (id SERIAL PRIMARY KEY);\nINSERT INTO gadgets (name) VALUES ('alpha');"

	if err := os.WriteFile(path, []byte(sql), 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	// The INSERT statement above is invalid because the table does not have a name column.
	if err := sqlseed.SeedFromFile(conn, path); err == nil {
		t.Fatalf("expected error when executing invalid sql")
	}

	if conn.Sql().Migrator().HasTable("gadgets") {
		t.Fatalf("expected transaction rollback to drop gadgets table")
	}
}

func setupPostgresConnection(t *testing.T) (*database.Connection, func()) {
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

	cleanup := func() {
		if err := conn.Ping(); err == nil {
			conn.Close()
		}

		if err := pg.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return conn, cleanup
}
