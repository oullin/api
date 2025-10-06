package sqlseed_test

import (
	"context"
	"fmt"
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

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "CREATE TABLE widgets (id SERIAL PRIMARY KEY, name TEXT NOT NULL);\nINSERT INTO widgets (name) VALUES ('alpha'), ('beta');")

	if err := sqlseed.SeedFromFile(conn, fileName); err != nil {
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
	fileName := writeStorageFile(t, withSuffix(t, ".txt"), "SELECT 1;")

	err := sqlseed.SeedFromFile(nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Fatalf("expected extension error, got %v", err)
	}
}

func TestSeedFromFileRejectsAbsolutePath(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	absPath, err := filepath.Abs(filepath.Join("storage", "sql", fileName))
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	err = sqlseed.SeedFromFile(nil, absPath)
	if err == nil || !strings.Contains(err.Error(), "absolute file paths") {
		t.Fatalf("expected absolute path error, got %v", err)
	}
}

func TestSeedFromFileRejectsTraversal(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	err := sqlseed.SeedFromFile(nil, filepath.Join("..", fileName))
	if err == nil || !strings.Contains(err.Error(), "within") {
		t.Fatalf("expected traversal error, got %v", err)
	}
}

func TestSeedFromFileFailsWhenFileMissing(t *testing.T) {
	fileName := withSuffix(t, "_missing.sql")

	err := sqlseed.SeedFromFile(nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "read file") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestSeedFromFileFailsWhenFileEmpty(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "   \n\t")

	err := sqlseed.SeedFromFile(nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected empty file error, got %v", err)
	}
}

func TestSeedFromFileRequiresConnection(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	err := sqlseed.SeedFromFile(nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "connection") {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func TestSeedFromFileRollsBackOnFailure(t *testing.T) {
	conn, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "CREATE TABLE gadgets (id SERIAL PRIMARY KEY);\nINSERT INTO gadgets (name) VALUES ('alpha');")

	// The INSERT statement above is invalid because the table does not have a name column.
	if err := sqlseed.SeedFromFile(conn, fileName); err == nil {
		t.Fatalf("expected error when executing invalid sql")
	}

	if conn.Sql().Migrator().HasTable("gadgets") {
		t.Fatalf("expected transaction rollback to drop gadgets table")
	}
}

func writeStorageFile(t *testing.T, name, contents string) string {
	t.Helper()

	dir := filepath.Join("storage", "sql")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create storage dir: %v", err)
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write storage file: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remove storage file: %v", err)
		}
	})

	return name
}

func withSuffix(t *testing.T, suffix string) string {
	t.Helper()

	base := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_")
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.ReplaceAll(base, ":", "_")

	return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), suffix)
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
