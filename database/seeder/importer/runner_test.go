package importer_test

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
	"github.com/oullin/database/seeder/importer"
	"github.com/oullin/metal/env"
)

func TestSeedFromFileExecutesStatements(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "CREATE TABLE widgets (id SERIAL PRIMARY KEY, name TEXT NOT NULL);\nINSERT INTO widgets (name) VALUES ('alpha'), ('beta');")

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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

	err := importer.SeedFromFile(nil, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Fatalf("expected extension error, got %v", err)
	}
}

func TestSeedFromFileRequiresEnvironment(t *testing.T) {
	conn, _, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	err := importer.SeedFromFile(conn, nil, fileName)
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

	err = importer.SeedFromFile(nil, nil, absPath)
	if err == nil || !strings.Contains(err.Error(), "absolute file paths") {
		t.Fatalf("expected absolute path error, got %v", err)
	}
}

func TestSeedFromFileRejectsTraversal(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	err := importer.SeedFromFile(nil, nil, filepath.Join("..", fileName))
	if err == nil || !strings.Contains(err.Error(), "within") {
		t.Fatalf("expected traversal error, got %v", err)
	}
}

func TestSeedFromFileFailsWhenFileMissing(t *testing.T) {
	fileName := withSuffix(t, "_missing.sql")

	err := importer.SeedFromFile(nil, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "read file") {
		t.Fatalf("expected read error, got %v", err)
	}
}

func TestSeedFromFileFailsWhenFileEmpty(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "   \n\t")

	err := importer.SeedFromFile(nil, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "is empty") {
		t.Fatalf("expected empty file error, got %v", err)
	}
}

func TestSeedFromFileRejectsNonUTF8Contents(t *testing.T) {
	fileName := writeStorageBytes(t, withSuffix(t, ".sql"), []byte{0xff, 0xfe, 0xfd})

	err := importer.SeedFromFile(nil, nil, fileName)
	if err == nil || !strings.Contains(err.Error(), "non-UTF-8") {
		t.Fatalf("expected non-UTF-8 error, got %v", err)
	}
}

func TestSeedFromFileRequiresConnection(t *testing.T) {
	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "SELECT 1;")

	err := importer.SeedFromFile(nil, testEnvironment(), fileName)
	if err == nil || !strings.Contains(err.Error(), "connection") {
		t.Fatalf("expected connection error, got %v", err)
	}
}

func TestSeedFromFileRollsBackOnFailure(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), "CREATE TABLE gadgets (id SERIAL PRIMARY KEY);\nINSERT INTO gadgets (name) VALUES ('alpha');")

	// The INSERT statement above is invalid because the table does not have a name column.
	err := importer.SeedFromFile(conn, environment, fileName)
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
		"3\tcontains\\\\.",
		"\\.",
		"",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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

	if len(rows) != 3 {
		t.Fatalf("expected 3 supplies, got %d", len(rows))
	}

	if rows[0].ID != 1 || rows[0].Name != "bolts" {
		t.Fatalf("unexpected first row: %+v", rows[0])
	}

	if rows[1].ID != 2 || rows[1].Name != "washers" {
		t.Fatalf("unexpected second row: %+v", rows[1])
	}

	if rows[2].ID != 3 || rows[2].Name != "contains\\." {
		t.Fatalf("unexpected third row: %+v", rows[2])
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

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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
		"WITH metadata AS (SELECT NOW() AS generated_at) INSERT INTO public.api_keys (id, uuid, account_name, public_key, secret_key, created_at, updated_at, deleted_at) SELECT 2, '00000000-0000-0000-0000-000000000202', 'with-account', '\\x03', '\\x04', generated_at, generated_at, NULL FROM metadata;",
		"COPY public.api_key_signatures (id, uuid, api_key_id, signature, max_tries, current_tries, expires_at, expired_at, origin, created_at, updated_at, deleted_at) FROM stdin;",
		copyRow,
		"\\.",
		"",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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
	assertCount("api_keys", 2)
	assertCount("api_key_signatures", 1)

	var nextVal int64
	if err := conn.Sql().Raw("SELECT nextval('public.api_keys_id_seq')").Scan(&nextVal).Error; err != nil {
		t.Fatalf("nextval sequence: %v", err)
	}

	if nextVal != 100 {
		t.Fatalf("expected nextval to be 100 after seeding API keys, got %d", nextVal)
	}
}

func TestSeedFromFileSkipsDropSequenceForExcludedTables(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	sequenceDefinition := strings.Join([]string{
		"CREATE SEQUENCE public.api_keys_id_seq",
		"    AS bigint",
		"    START WITH 1",
		"    INCREMENT BY 1",
		"    NO MINVALUE",
		"    NO MAXVALUE",
		"    CACHE 1;",
	}, "\n")

	contents := strings.Join([]string{
		"DROP SEQUENCE IF EXISTS public.api_keys_id_seq;",
		sequenceDefinition,
		"ALTER SEQUENCE public.api_keys_id_seq OWNER TO test;",
		"ALTER SEQUENCE public.api_keys_id_seq OWNED BY public.api_keys.id;",
		"SELECT pg_catalog.setval('public.api_keys_id_seq', 42, true);",
		"INSERT INTO users (uuid, first_name, last_name, username, email, password_hash, public_token) VALUES ('00000000-0000-0000-0000-000000000111', 'Jane', 'Doe', 'janedoe', 'jane@example.com', 'hash', 'token');",
		"",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	var nextVal int64
	if err := conn.Sql().Raw("SELECT nextval('public.api_keys_id_seq')").Scan(&nextVal).Error; err != nil {
		t.Fatalf("nextval sequence: %v", err)
	}

	if nextVal != 43 {
		t.Fatalf("expected nextval to be 43 after recreating sequence, got %d", nextVal)
	}
}

func TestSeedFromFileSkipsDuplicateConstraintAdds(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"ALTER TABLE ONLY public.categories ADD CONSTRAINT categories_name_key UNIQUE (name);",
		"",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}
}

func TestSeedFromFileSkipsDuplicatePrimaryKeyAdds(t *testing.T) {
	conn, environment, cleanup := setupPostgresConnection(t)
	t.Cleanup(cleanup)

	contents := strings.Join([]string{
		"CREATE TABLE duplicate_pk (id BIGINT PRIMARY KEY, name TEXT NOT NULL);",
		"ALTER TABLE ONLY public.duplicate_pk ADD CONSTRAINT duplicate_pk_pkey PRIMARY KEY (id);",
		"INSERT INTO duplicate_pk (id, name) VALUES (1, 'alpha');",
		"",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	var count int64
	if err := conn.Sql().Table("duplicate_pk").Count(&count).Error; err != nil {
		t.Fatalf("count duplicate_pk: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 row in duplicate_pk, got %d", count)
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

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("first seed: %v", err)
	}

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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
		"INSERT INTO categories (uuid, name, slug)",
		"VALUES ('00000000-0000-0000-0000-00000000c001', 'Tech', 'tech');",
	}, "\n")

	fileName := writeStorageFile(t, withSuffix(t, ".sql"), contents)

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("seed from file: %v", err)
	}

	var count int64
	if err := conn.Sql().Table("categories").Count(&count).Error; err != nil {
		t.Fatalf("count categories: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 category, got %d", count)
	}

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
		t.Fatalf("second seed from file: %v", err)
	}

	count = 0
	if err := conn.Sql().Table("categories").Count(&count).Error; err != nil {
		t.Fatalf("count categories after reseed: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 category after reseed, got %d", count)
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

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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

	if err := importer.SeedFromFile(conn, environment, fileName); err != nil {
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

	err := importer.SeedFromFile(conn, environment, fileName)
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

	if os.Getenv("IMPORTER_SKIP_INTEGRATION") == "1" || os.Getenv("SQLSEED_SKIP_INTEGRATION") == "1" {
		t.Skip("importer integration tests disabled via IMPORTER_SKIP_INTEGRATION or SQLSEED_SKIP_INTEGRATION")
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
