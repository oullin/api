package importer

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/oullin/database"
	"github.com/oullin/metal/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestExecuteStatementsCommitsStatements(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SET LOCAL session_replication_role = 'replica'")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_1")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO widgets VALUES (1)")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("RELEASE SAVEPOINT importer_sp_1")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err := executeStatements(context.Background(), conn, []statement{{sql: "INSERT INTO widgets VALUES (1)"}}, executeOptions{disableConstraints: true})
	if err != nil {
		t.Fatalf("executeStatements unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecuteStatementsRollsBackOnError(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SET LOCAL session_replication_role = 'replica'")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_1")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO widgets VALUES (1)")).
		WillReturnError(errors.New("boom"))
	mock.ExpectExec(regexp.QuoteMeta("ROLLBACK TO SAVEPOINT importer_sp_1")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := executeStatements(context.Background(), conn, []statement{{sql: "INSERT INTO widgets VALUES (1)"}}, executeOptions{disableConstraints: true})
	if err == nil || !strings.Contains(err.Error(), "statement 1") {
		t.Fatalf("expected error about statement 1, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecuteStatementsSkipsHandledErrors(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_1")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations VALUES (42, false)")).
		WillReturnError(errors.New("ERROR: duplicate key value violates unique constraint \"schema_migrations_pkey\" (SQLSTATE 23505)"))
	mock.ExpectExec(regexp.QuoteMeta("ROLLBACK TO SAVEPOINT importer_sp_1")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("RELEASE SAVEPOINT importer_sp_1")).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	output := captureStderr(t, func() {
		err := executeStatements(context.Background(), conn, []statement{{sql: "INSERT INTO schema_migrations VALUES (42, false)"}}, executeOptions{})
		if err != nil {
			t.Fatalf("executeStatements unexpected error: %v", err)
		}
	})

	if !strings.Contains(output, "duplicate migration row") {
		t.Fatalf("expected skip reason in stderr, got %q", output)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecuteCopyInsertsRows(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectBegin()
	insert := regexp.QuoteMeta("INSERT INTO public.widgets (id, name) VALUES ($1, $2)")
	mock.ExpectExec(insert).WithArgs("1", "foo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(insert).WithArgs("2", "bar").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	stmt := statement{
		sql:      "COPY public.widgets (id, name) FROM stdin",
		copyData: []byte("1\tfoo\n2\tbar"),
		isCopy:   true,
	}

	if err := executeCopy(context.Background(), tx, stmt); err != nil {
		t.Fatalf("executeCopy unexpected error: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecuteCopyResolvesColumns(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT column_name FROM information_schema.columns").
		WithArgs("public", "widgets").
		WillReturnRows(sqlmock.NewRows([]string{"column_name"}).AddRow("id").AddRow("name"))
	insert := regexp.QuoteMeta("INSERT INTO public.widgets (id, name) VALUES ($1, $2)")
	mock.ExpectExec(insert).WithArgs("1", "foo").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	stmt := statement{
		sql:      "COPY public.widgets FROM stdin",
		copyData: []byte("1\tfoo"),
		isCopy:   true,
	}

	if err := executeCopy(context.Background(), tx, stmt); err != nil {
		t.Fatalf("executeCopy unexpected error: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveCopyColumnsErrorsOnEmptyMetadata(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT column_name FROM information_schema.columns").
		WithArgs("public", "widgets").
		WillReturnRows(sqlmock.NewRows([]string{"column_name"}))
	mock.ExpectRollback()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	_, err = resolveCopyColumns(context.Background(), tx, "public.widgets")
	if err == nil || !strings.Contains(err.Error(), "has no columns") {
		t.Fatalf("expected empty metadata error, got %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveCopyColumnsPropagatesQueryError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT column_name FROM information_schema.columns").
		WithArgs("public", "widgets").
		WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	_, err = resolveCopyColumns(context.Background(), tx, "public.widgets")
	if err == nil || !strings.Contains(err.Error(), "query failed") {
		t.Fatalf("expected query error, got %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPrepareDatabaseReturnsTruncateError(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	environment := &env.Environment{App: env.AppEnvironment{Type: "local"}}
	expectTruncateTables(t, mock, nil, "api_key_signatures", errors.New("truncate boom"))

	err := prepareDatabase(context.Background(), conn, environment)
	if err == nil || !strings.Contains(err.Error(), "truncate database") {
		t.Fatalf("expected truncate error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPrepareDatabaseExecutesMigrations(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "database", "infra", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations dir: %v", err)
	}

	contents := strings.Join([]string{
		"CREATE TABLE foo (id INT);",
		"INSERT INTO foo (id) VALUES (1);",
		"",
	}, "\n")

	if err := os.WriteFile(filepath.Join(migrationsDir, "002_test.up.sql"), []byte(contents), 0o600); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	expectTruncateTables(t, mock, nil, "", nil)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE foo (id INT)")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("RELEASE SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_2")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO foo (id) VALUES (1)")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("RELEASE SAVEPOINT importer_sp_2")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	environment := &env.Environment{App: env.AppEnvironment{Type: "local"}}
	if err := prepareDatabase(context.Background(), conn, environment); err != nil {
		t.Fatalf("prepareDatabase unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSeedFromFileExecutesSQLStatements(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "database", "infra", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations dir: %v", err)
	}

	storageDir := filepath.Join(tempDir, "storage", "sql")
	if err := os.MkdirAll(storageDir, 0o755); err != nil {
		t.Fatalf("mkdir storage dir: %v", err)
	}

	migrationContents := strings.Join([]string{
		"CREATE TABLE foo (id INT);",
		"",
	}, "\n")

	if err := os.WriteFile(filepath.Join(migrationsDir, "003_test.up.sql"), []byte(migrationContents), 0o600); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	seedContents := strings.Join([]string{
		"INSERT INTO foo (id) VALUES (99);",
		"",
	}, "\n")

	if err := os.WriteFile(filepath.Join(storageDir, "seed.sql"), []byte(seedContents), 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	expectTruncateTables(t, mock, nil, "", nil)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE foo (id INT)")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("RELEASE SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SET LOCAL session_replication_role = 'replica'")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO foo (id) VALUES (99)")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("RELEASE SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	environment := &env.Environment{App: env.AppEnvironment{Type: "local"}}
	if err := SeedFromFile(conn, environment, "seed.sql"); err != nil {
		t.Fatalf("SeedFromFile unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecuteStatementsSkipsExcludedTable(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SET LOCAL session_replication_role = 'replica'")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	stmt := statement{sql: "INSERT INTO public.api_keys VALUES (1)"}
	err := executeStatements(context.Background(), conn, []statement{stmt}, executeOptions{
		disableConstraints: true,
		skipTables:         excludedSeedTables,
	})
	if err != nil {
		t.Fatalf("executeStatements unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestExecuteStatementsRollbackFailure(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO widgets VALUES (1)")).WillReturnError(errors.New("boom"))
	mock.ExpectExec(regexp.QuoteMeta("ROLLBACK TO SAVEPOINT importer_sp_1")).WillReturnError(errors.New("rollback failed"))
	mock.ExpectRollback()

	stmt := statement{sql: "INSERT INTO widgets VALUES (1)"}
	err := executeStatements(context.Background(), conn, []statement{stmt}, executeOptions{})
	if err == nil || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("expected combined rollback error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRunMigrationsExecutesStatements(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "database", "infra", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations dir: %v", err)
	}

	contents := strings.Join([]string{
		"CREATE TABLE foo (id INT);",
		"INSERT INTO foo (id) VALUES (1);",
		"",
	}, "\n")

	if err := os.WriteFile(filepath.Join(migrationsDir, "001_test.up.sql"), []byte(contents), 0o600); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE foo (id INT)")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("RELEASE SAVEPOINT importer_sp_1")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("SAVEPOINT importer_sp_2")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO foo (id) VALUES (1)")).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("RELEASE SAVEPOINT importer_sp_2")).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	if err := runMigrations(context.Background(), conn); err != nil {
		t.Fatalf("runMigrations unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRunMigrationsReturnsParseError(t *testing.T) {
	conn, mock, sqlDB := newMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	tempDir := t.TempDir()
	migrationsDir := filepath.Join(tempDir, "database", "infra", "migrations")
	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		t.Fatalf("mkdir migrations dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(migrationsDir, "004_bad.up.sql"), []byte("CREATE TABLE broken (id INT"), 0o600); err != nil {
		t.Fatalf("write migration: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	err = runMigrations(context.Background(), conn)
	if err == nil || !strings.Contains(err.Error(), "parse migration") {
		t.Fatalf("expected parse error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRollbackToSavepointReturnsError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer sqlDB.Close()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("ROLLBACK TO SAVEPOINT failing")).WillReturnError(errors.New("boom"))
	mock.ExpectRollback()

	tx, err := sqlDB.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	err = rollbackToSavepoint(context.Background(), tx, "failing")
	if err == nil || !strings.Contains(err.Error(), "rollback savepoint failing") {
		t.Fatalf("expected rollback error, got %v", err)
	}

	if rbErr := tx.Rollback(); rbErr != nil {
		t.Fatalf("rollback: %v", rbErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func expectTruncateTables(t *testing.T, mock sqlmock.Sqlmock, existing map[string]bool, errorTable string, errorToReturn error) {
	t.Helper()

	query := regexp.QuoteMeta("SELECT count(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND table_type = $2")
	tables := database.GetSchemaTables()

	for i := len(tables) - 1; i >= 0; i-- {
		table := tables[i]
		exists := true
		if existing != nil {
			if val, ok := existing[table]; ok {
				exists = val
			}
		}

		count := int64(0)
		if exists {
			count = 1
		}

		mock.ExpectQuery(query).
			WithArgs(table, "BASE TABLE").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))

		if !exists {
			continue
		}

		stmt := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table)
		expectation := mock.ExpectExec(regexp.QuoteMeta(stmt))
		if table == errorTable && errorToReturn != nil {
			expectation.WillReturnError(errorToReturn)
		} else {
			expectation.WillReturnResult(sqlmock.NewResult(0, 0))
		}
	}
}

func newMockConnection(t *testing.T) (*database.Connection, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	})

	gdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	conn := database.NewConnectionFromGorm(gdb)
	return conn, mock, sqlDB
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	os.Stderr = w
	defer func() {
		os.Stderr = original
	}()

	done := make(chan struct{})
	var buf bytes.Buffer

	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close write pipe: %v", err)
	}

	<-done
	return buf.String()
}
