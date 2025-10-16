package database_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

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

	conn, err := database.NewConnection(e)
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
	sqlDB := sql.OpenDB(newErrCloseConnector(errors.New("boom")))

	db, err := gorm.Open(noopDialector{connPool: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}

	conn := database.NewConnectionFromGorm(db)

	if ok := conn.Close(); ok {
		t.Fatalf("expected close to report failure")
	}
}

type noopDialector struct {
	connPool gorm.ConnPool
}

func (d noopDialector) Name() string {
	return "noop"
}

func (d noopDialector) Initialize(db *gorm.DB) error {
	db.Config.ConnPool = d.connPool
	return nil
}

func (noopDialector) Migrator(*gorm.DB) gorm.Migrator {
	return nil
}

func (noopDialector) DataTypeOf(*schema.Field) string {
	return ""
}

func (noopDialector) DefaultValueOf(*schema.Field) clause.Expression {
	return clause.Expr{}
}

func (noopDialector) BindVarTo(writer clause.Writer, _ *gorm.Statement, _ interface{}) {
	writer.WriteByte('?')
}

func (noopDialector) QuoteTo(writer clause.Writer, s string) {
	writer.WriteString(s)
}

func (noopDialector) Explain(sql string, _ ...interface{}) string {
	return sql
}

type errCloseConnector struct {
	driver.Connector
	closeErr error
}

func newErrCloseConnector(closeErr error) *errCloseConnector {
	return &errCloseConnector{
		Connector: &noopConnector{},
		closeErr:  closeErr,
	}
}

func (c *errCloseConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return c.Connector.Connect(ctx)
}

func (c *errCloseConnector) Driver() driver.Driver {
	return c.Connector.Driver()
}

func (c *errCloseConnector) Close() error {
	return c.closeErr
}

type noopConnector struct{}

func (*noopConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return &noopConn{}, nil
}

func (*noopConnector) Driver() driver.Driver {
	return noopDriver{}
}

type noopDriver struct{}

func (noopDriver) Open(string) (driver.Conn, error) {
	return &noopConn{}, nil
}

type noopConn struct{}

func (*noopConn) Prepare(string) (driver.Stmt, error) {
	return noopStmt{}, nil
}

func (*noopConn) Close() error { return nil }

func (*noopConn) Begin() (driver.Tx, error) { return noopTx{}, nil }

type noopStmt struct{}

func (noopStmt) Close() error { return nil }

func (noopStmt) NumInput() int { return 0 }

func (noopStmt) Exec([]driver.Value) (driver.Result, error) { return noopResult{}, nil }

func (noopStmt) Query([]driver.Value) (driver.Rows, error) { return noopRows{}, nil }

type noopTx struct{}

func (noopTx) Commit() error { return nil }

func (noopTx) Rollback() error { return nil }

type noopResult struct{}

func (noopResult) LastInsertId() (int64, error) { return 0, nil }

func (noopResult) RowsAffected() (int64, error) { return 0, nil }

type noopRows struct{}

func (noopRows) Columns() []string { return nil }

func (noopRows) Close() error { return nil }

func (noopRows) Next(dest []driver.Value) error { return io.EOF }

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
