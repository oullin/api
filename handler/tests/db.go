package handlertests

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/oullin/database"
	"github.com/oullin/metal/env"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// NewTestDB starts a PostgreSQL test container, runs migrations, and seeds a default user.
// It returns the database connection and the created user.
func NewTestDB(t *testing.T) (*database.Connection, database.User) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed")
	}
	if err := exec.Command("docker", "ps").Run(); err != nil {
		t.Skip("docker not running")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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
	t.Cleanup(func() { pg.Terminate(context.Background()) })

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
		t.Fatalf("new connection: %v", err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := conn.Sql().AutoMigrate(&database.User{}, &database.Post{}, &database.Category{}, &database.Tag{}, &database.PostCategory{}, &database.PostTag{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	author := database.User{
		UUID:         uuid.NewString(),
		Username:     "user",
		FirstName:    "F",
		LastName:     "L",
		Email:        "u@example.com",
		PasswordHash: "x",
	}
	if err := conn.Sql().Create(&author).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return conn, author
}
