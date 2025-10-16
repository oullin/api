package database

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/oullin/metal/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestTruncateExecuteSkipsMissingTables(t *testing.T) {
	conn, mock, sqlDB := newTruncateMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	existing := allTables(false)
	expectTruncateCalls(t, mock, existing, nil)

	truncate := NewTruncate(conn, &env.Environment{App: env.AppEnvironment{Type: "local"}})
	if err := truncate.Execute(); err != nil {
		t.Fatalf("Execute unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTruncateExecuteSkipsUndefinedRelationErrors(t *testing.T) {
	conn, mock, sqlDB := newTruncateMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	existing := allTables(false)
	existing["users"] = true
	execErrors := map[string]error{
		"users": errors.New("ERROR: relation \"users\" does not exist (SQLSTATE 42P01)"),
	}

	expectTruncateCalls(t, mock, existing, execErrors)

	truncate := NewTruncate(conn, &env.Environment{App: env.AppEnvironment{Type: "local"}})
	if err := truncate.Execute(); err != nil {
		t.Fatalf("Execute unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTruncateExecuteAggregatesErrors(t *testing.T) {
	conn, mock, sqlDB := newTruncateMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	existing := allTables(false)
	existing["users"] = true
	execErrors := map[string]error{
		"users": errors.New("truncate boom"),
	}

	expectTruncateCalls(t, mock, existing, execErrors)

	truncate := NewTruncate(conn, &env.Environment{App: env.AppEnvironment{Type: "local"}})
	err := truncate.Execute()
	if err == nil || !regexp.MustCompile(`truncate table users`).MatchString(err.Error()) {
		t.Fatalf("expected error about users table, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTruncateExecutePanicsInProduction(t *testing.T) {
	conn, _, sqlDB := newTruncateMockConnection(t)
	t.Cleanup(func() { _ = sqlDB.Close() })

	truncate := NewTruncate(conn, &env.Environment{App: env.AppEnvironment{Type: "production"}})

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when truncating production")
		}
	}()

	_ = truncate.Execute()
}

func newTruncateMockConnection(t *testing.T) (*Connection, sqlmock.Sqlmock, *sql.DB) {
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

	conn := NewConnectionFromGorm(gdb)
	return conn, mock, sqlDB
}

func allTables(value bool) map[string]bool {
	tables := GetSchemaTables()
	result := make(map[string]bool, len(tables))
	for _, table := range tables {
		result[table] = value
	}
	return result
}

func expectTruncateCalls(t *testing.T, mock sqlmock.Sqlmock, existing map[string]bool, execErrors map[string]error) {
	t.Helper()

	query := regexp.QuoteMeta("SELECT count(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND table_type = $2")
	tables := GetSchemaTables()

	for i := len(tables) - 1; i >= 0; i-- {
		table := tables[i]
		exists := true
		if existing != nil {
			exists = existing[table]
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
		if err, ok := execErrors[table]; ok {
			expectation.WillReturnError(err)
		} else {
			expectation.WillReturnResult(sqlmock.NewResult(0, 0))
		}
	}
}
