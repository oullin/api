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
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "CREATE TABLE widgets (id SERIAL PRIMARY KEY, name TEXT NOT NULL);\nINSERT INTO widgets (name) VALUES ('alpha'), ('beta');")

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
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

	err := sqlseed.SeedFromFile(nil, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Fatalf("expected extension error, got %v", err)
	}
}

func TestSeedFromFileRequiresEnvironment(t *testing.T) {
	conn, _, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	err := sqlseed.SeedFromFile(conn, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "environment is required") {
		t.Fatalf("expected environment error, got %v", err)
	}
}

func TestSeedFromFileRejectsAbsolutePath(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	absPath, err := filepath.Abs(filepath.Join("storage", "sql", fileName))
	if err != nil {
		t.Fatalf("abs path: %v", err)
	}

	err = sqlseed.SeedFromFile(nil, nil, absPath)
	if err == nil || !strings.Contains(err.Error(), "absolute file paths") {
		t.Fatalf("expected absolute path error, got %v", err)
	}
}

func TestSeedFromFileRejectsTraversal(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	err := sqlseed.SeedFromFile(nil, nil, filepath.Join("..", fileName))
	if err == nil || !strings.Contains(err.Error(), "within") {
		t.Fatalf("expected traversal error, got %v", err)
	}
}

func TestSeedFromFileFailsWhenFileMissing(t *testing.T) {
	fileName := withSuffix(t, "_missing.sql")

	err := sqlseed.SeedFromFile(nil, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "read file") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestSeedFromFileFailsWhenFileEmpty(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "   \n\t")

	err := sqlseed.SeedFromFile(nil, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("expected empty file error, got %v", err)
	}
}

func TestSeedFromFileRejectsNonUTF8Contents(t *testing.T) {
	fileName := writeStorageBytes(t, withSuffix(t, ".sql"), []byte{0xff, 0xfe, 0xfd})

	err := sqlseed.SeedFromFile(nil, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "non-UTF-8") {
		t.Fatalf("expected non-UTF-8 error, got %v", err)
	}
}

func TestSeedFromFileRequiresConnection(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	err := sqlseed.SeedFromFile(nil, testEnvironment(), fileName)
	if err == nil || !strings.Contains(err.Error(), "connection") {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func TestSeedFromFileRollsBackOnFailure(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "CREATE TABLE gadgets (id SERIAL PRIMARY KEY);\nINSERT INTO gadgets (name) VALUES ('alpha');")

	// The INSERT statement above is invalid because the table does not have a name column.
	err := sqlseed.SeedFromFile(conn, environment, fileName)
	if err == nil {
		t.Fatalf("expected error when executing invalid sql")
	}

	if !strings.Contains(err.Error(), "statement 2") {
		t.Fatalf("expected error to identify failing statement, got %v", err)
	}

	if !strings.Contains(err.Error(), "INSERT INTO gadgets") {
		t.Fatalf("expected error to include statement preview, got %v", err)
	}

	if conn.Sql().Migrator().HasTable("gadgets") {
		t.Fatalf("expected transaction rollback to drop gadgets table")
	}
}

func TestSeedFromFileSupportsCopyFromStdin(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"CREATE TABLE supplies (id INTEGER PRIMARY KEY, name TEXT NOT NULL);",
		"COPY supplies (id, name) FROM stdin;",
		"1\tbolts",
		"2\twashers",
		"\\.",
		"",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	type supply struct {
		ID   int
		Name string
	}

	var rows []supply
	if err := conn.Sql().Table("supplies").Order("id").Find(&rows).Error; err != nil {
		t.Fatalf("query supplies: %v", err)
	}

	if len(rows) != 2 {
		t.Fatalf("expected 2 supplies, got %d", len(rows))
	}

	if rows[0].ID != 1 || rows[0].Name != "bolts" {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}

	if rows[1].ID != 2 || rows[1].Name != "washers" {
		t.Fatalf("unexpected second row: %+v", rows[1])
	}
}

func TestSeedFromFileLoadsDataOutOfConstraintOrder(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"CREATE TABLE parents (id BIGINT PRIMARY KEY, name TEXT NOT NULL);",
		"CREATE TABLE children (id BIGINT PRIMARY KEY, parent_id BIGINT NOT NULL REFERENCES parents(id), name TEXT NOT NULL);",
		"INSERT INTO children (id, parent_id, name) VALUES (1, 42, 'child-before-parent');",
		"INSERT INTO parents (id, name) VALUES (42, 'parent-later');",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	type child struct {
		ID       int64
		ParentID int64
		Name     string
	}

	var rows []child
	if err := conn.Sql().Table("children").Find(&rows).Error; err != nil {
		t.Fatalf("query children: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 child, got %d", len(rows))
	}

	if rows[0].ParentID != 42 {
		t.Fatalf("expected parent id 42, got %d", rows[0].ParentID)
	}

	var parentCount int64
	if err := conn.Sql().Table("parents").Where("id = ?", 42).Count(&parentCount).Error; err != nil {
		t.Fatalf("count parents: %v", err)
	}

	if parentCount != 1 {
		t.Fatalf("expected parent to exist, got %d", parentCount)
	}
}

func TestSeedFromFileSkipsExcludedTables(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	copyRow := strings.Join([]string{
		"1",
		"00000000-0000-0000-0000-000000000301",
		"1",
		"\\xdeadbeef",
		"5",
		"1",
		"\\N",
		"\\N",
		"excluded",
		"2024-01-02 03:04:05",
		"2024-01-02 03:04:05",
		"\\N",
	}, "\t")

	contents := strings.Join([]string{
		"INSERT INTO users (uuid, first_name, last_name, username, email, password_hash, public_token) VALUES ('00000000-0000-0000-0000-000000000101', 'Alice', 'Smith', 'asmith', 'alice@example.com', 'hash', 'token');",
		"INSERT INTO public.api_keys (id, uuid, account_name, public_key, secret_key, created_at, updated_at, deleted_at) VALUES (1, '00000000-0000-0000-0000-000000000201', 'demo-account', '\\x01', '\\x02', NOW(), NOW(), NULL);",
		"SELECT pg_catalog.setval('public.api_keys_id_seq', 99, true);",
		"COPY public.api_key_signatures (id, uuid, api_key_id, signature, max_tries, current_tries, expires_at, expired_at, origin, created_at, updated_at, deleted_at) FROM stdin;",
		copyRow,
		"\\.",
		"",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	assertCount := func(table string, expected int64) {
		t.Helper()
		var count int64
		if err := conn.Sql().Table(table).Count(&count).Error; err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != expected {
			t.Fatalf("expected %d rows in %s, got %d", expected, table, count)
		}
	}

	assertCount("users", 1)
	assertCount("api_keys", 0)
	assertCount("api_key_signatures", 0)
}

func TestSeedFromFileSkipsDropSequenceForExcludedTables(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"INSERT INTO users (uuid, first_name, last_name, username, email, password_hash, public_token) VALUES ('00000000-0000-0000-0000-000000000111', 'Jane', 'Doe', 'janedoe', 'jane@example.com', 'hash', 'token');",
		"DROP SEQUENCE public.api_keys_id_seq;",
		"",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	var nextVal int64
	if err := conn.Sql().Raw("SELECT nextval('public.api_keys_id_seq')").Scan(&nextVal).Error; err != nil {
		t.Fatalf("nextval sequence: %v", err)
	}

	var userCount int64
	if err := conn.Sql().Table("users").Count(&userCount).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}

	if userCount != 1 {
		t.Fatalf("expected 1 user after seeding, got %d", userCount)
	}
}

func TestSeedFromFileSkipsDuplicateCreates(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"CREATE TABLE widgets (id SERIAL PRIMARY KEY, name TEXT NOT NULL);",
		"INSERT INTO widgets (name) VALUES ('alpha'), ('beta');",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("second seed: %v", err)
	}

	var count int64
	if err := conn.Sql().Table("widgets").Count(&count).Error; err != nil {
		t.Fatalf("count widgets: %v", err)
	}

	if count != 4 {
		t.Fatalf("expected 4 widgets after reseeding, got %d", count)
	}
}

func TestSeedFromFileSkipsMissingOwnerRole(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"CREATE TABLE owner_change (id SERIAL PRIMARY KEY, note TEXT NOT NULL);",
		"ALTER TABLE owner_change OWNER TO missing_role;",
		"INSERT INTO owner_change (note) VALUES ('ok');",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	type ownerRow struct {
		Note string
	}

	var rows []ownerRow
	if err := conn.Sql().Table("owner_change").Find(&rows).Error; err != nil {
		t.Fatalf("query owner_change: %v", err)
	}

	if len(rows) != 1 || rows[0].Note != "ok" {
		t.Fatalf("unexpected owner_change rows: %+v", rows)
	}
}

func TestSeedFromFileSkipsMissingRelationAlter(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"ALTER TABLE missing_relation DROP CONSTRAINT missing_relation_pkey;",
		"CREATE TABLE relation_keep (id SERIAL PRIMARY KEY, note TEXT NOT NULL);",
		"INSERT INTO relation_keep (note) VALUES ('ok');",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	type relationRow struct {
		Note string
	}

	var rows []relationRow
	if err := conn.Sql().Table("relation_keep").Find(&rows).Error; err != nil {
		t.Fatalf("query relation_keep: %v", err)
	}

	if len(rows) != 1 || rows[0].Note != "ok" {
		t.Fatalf("unexpected relation_keep rows: %+v", rows)
	}
}

func TestSeedFromFileRunsMigrations(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"INSERT INTO api_keys (uuid, account_name, public_key, secret_key)",
		"VALUES ('00000000-0000-0000-0000-000000000001', 'example', decode('68656c6c6f', 'hex'), decode('736563726574', 'hex'));",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	var count int64
	if err := conn.Sql().Table("api_keys").Count(&count).Error; err != nil {
		t.Fatalf("count api_keys: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 api key, got %d", count)
	}

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("second seed from file: %v", err)
	}

	count = 0
	if err := conn.Sql().Table("api_keys").Count(&count).Error; err != nil {
		t.Fatalf("count api_keys after reseed: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 api key after reseed, got %d", count)
	}
}

func TestSeedFromFileAllowsTrailingComment(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"CREATE TABLE comment_samples (id SERIAL PRIMARY KEY);",
		"INSERT INTO comment_samples DEFAULT VALUES;",
		"-- trailing comment without terminating semicolon",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	var count int64
	if err := conn.Sql().Table("comment_samples").Count(&count).Error; err != nil {
		t.Fatalf("count comment_samples: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestSeedFromFileSkipsMetaCommands(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"CREATE TABLE meta_samples (id SERIAL PRIMARY KEY, name TEXT NOT NULL);",
		"\\unrestrict RxSYF91vrSQYWEkG1ncg4fXRoYz64lllqFyU6He6bfAOnQdm2YZx8nLWBqOC8XK",
		"INSERT INTO meta_samples (name) VALUES ('alpha');",
		"\\connect - oullin",
		"INSERT INTO meta_samples (name) VALUES ('beta');",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := sqlseed.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	var count int64
	if err := conn.Sql().Table("meta_samples").Count(&count).Error; err != nil {
		t.Fatalf("count meta_samples: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected 2 rows, got %d", count)
	}
}

func TestSeedFromFileReportsUnterminatedStatementDetails(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"CREATE TABLE debug_statements (id SERIAL PRIMARY KEY)",
		"-- missing semicolon should trigger parser diagnostics",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	err := sqlseed.SeedFromFile(conn, environment, fileName)
	if err == nil {
		t.Fatalf("expected parse error for unterminated statement")
	}

	if !strings.Contains(err.Error(), "unterminated statement") {
		t.Fatalf("expected unterminated statement error, got %v", err)
	}

	if !strings.Contains(err.Error(), "line 1") {
		t.Fatalf("expected error to report line number, got %v", err)
	}

	if !strings.Contains(err.Error(), "debug_statements") {
		t.Fatalf("expected error to include statement preview, got %v", err)
	}
}

func testEnvironment() *env.Environment {
	return &env.Environment{App: env.AppEnvironment{Type: "local"}}
}

func writeStorageFile(t *testing.T, name, contents string) string {
	t.Helper()

	return writeStorageBytes(t, name, []byte(contents))
}

func writeStorageBytes(t *testing.T, name string, contents []byte) string {
	t.Helper()

	dir := filepath.Join("storage", "sql")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create storage dir: %v", err)
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
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

func setupPostgresConnection(t *testing.T) (*database.Connection, *env.Environment, func()) {
	t.Helper()

	if os.Getenv("SQLSEED_SKIP_INTEGRATION") == "1" {
		t.Skip("sqlseed integration tests disabled via SQLSEED_SKIP_INTEGRATION")
	}

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
		App: env.AppEnvironment{Type: "local"},
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

	return conn, e, cleanup
}
