package database_test

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/oullin/database"
	"github.com/oullin/database/repository"
	"github.com/oullin/metal/env"
)

func setupPostgresConnection(t *testing.T, models ...interface{}) (*database.Connection, func()) {
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

	if len(models) > 0 {
		if err := conn.Sql().AutoMigrate(models...); err != nil {
			t.Fatalf("migrate err: %v", err)
		}
	}

	cleanup := func() {
		if err := conn.Ping(); err == nil {
			conn.Close()
		}

		_ = pg.Terminate(context.Background())
	}

	t.Cleanup(cleanup)

	return conn, cleanup
}

func TestConnectionPingSuccess(t *testing.T) {
	conn, _ := setupPostgresConnection(t)

	if err := conn.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestConnectionPingReturnsErrorWhenPingFails(t *testing.T) {
	conn, _ := setupPostgresConnection(t)

	sqlDB, err := conn.Sql().DB()
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
	conn, _ := setupPostgresConnection(t)

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

	db, err := gorm.Open(gormpostgres.New(gormpostgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}), &gorm.Config{})
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
	conn, _ := setupPostgresConnection(t)

	if conn.Sql() == nil {
		t.Fatalf("expected sql to return underlying driver")
	}
}

func TestConnectionGetSessionEnablesQueryFields(t *testing.T) {
	conn, _ := setupPostgresConnection(t)

	session := conn.GetSession()
	if !session.QueryFields {
		t.Fatalf("expected session to enable query fields")
	}
}

func TestConnectionTransaction(t *testing.T) {
	conn, _ := setupPostgresConnection(t)

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
	conn, _ := setupPostgresConnection(t)

	expected := errors.New("boom")

	if err := conn.Transaction(func(tx *gorm.DB) error {
		return expected
	}); !errors.Is(err, expected) {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}

func TestApiKeysWithTestContainer(t *testing.T) {
	conn, _ := setupPostgresConnection(t, &database.APIKey{})

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
