package database_test

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/oullin/database"
)

func newSQLiteConnection(t *testing.T) (*database.Connection, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("unwrap sql db: %v", err)
	}

	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return database.NewConnectionFromGorm(db), db
}

func TestConnectionPingSuccess(t *testing.T) {
	conn, _ := newSQLiteConnection(t)

	if err := conn.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestConnectionPingReturnsErrorWhenPingFails(t *testing.T) {
	conn, db := newSQLiteConnection(t)

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("unwrap sql db: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close sql db: %v", err)
	}

	if err := conn.Ping(); err == nil {
		t.Fatalf("expected ping error after closing db")
	}
}

func TestConnectionCloseSuccess(t *testing.T) {
	conn, _ := newSQLiteConnection(t)

	if ok := conn.Close(); !ok {
		t.Fatalf("expected close to succeed")
	}
}

func TestConnectionCloseReturnsFalseOnError(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	mock.ExpectClose().WillReturnError(errors.New("boom"))

	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	conn := database.NewConnectionFromGorm(db)

	if ok := conn.Close(); ok {
		t.Fatalf("expected close to report failure")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("close expectations: %v", err)
	}
}

func TestConnectionSqlReturnsDriver(t *testing.T) {
	conn, db := newSQLiteConnection(t)

	if conn.Sql() != db {
		t.Fatalf("expected sql to return underlying driver")
	}
}

func TestConnectionGetSessionEnablesQueryFields(t *testing.T) {
	conn, _ := newSQLiteConnection(t)

	session := conn.GetSession()
	if !session.QueryFields {
		t.Fatalf("expected session to enable query fields")
	}
}

func TestConnectionTransaction(t *testing.T) {
	conn, _ := newSQLiteConnection(t)

	executed := false
	if err := conn.Transaction(func(tx *gorm.DB) error {
		executed = true
		return nil
	}); err != nil {
		t.Fatalf("transaction err: %v", err)
	}

	if !executed {
		t.Fatalf("expected callback to execute")
	}
}

func TestConnectionTransactionPropagatesError(t *testing.T) {
	conn, _ := newSQLiteConnection(t)

	expected := errors.New("boom")

	if err := conn.Transaction(func(tx *gorm.DB) error {
		return expected
	}); !errors.Is(err, expected) {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}
