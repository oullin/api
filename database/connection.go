package database

import (
	"database/sql"
	"fmt"
	"github.com/oullin/metal/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log/slog"
)

type Connection struct {
	url        string
	driverName string
	driver     *gorm.DB
	env        *env.Environment
}

func MakeConnection(env *env.Environment) (*Connection, error) {
	dbEnv := env.DB
	driver, err := gorm.Open(postgres.Open(dbEnv.GetDSN()), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return &Connection{
		driver:     driver,
		driverName: dbEnv.DriverName,
		env:        env,
	}, nil
}

func (c *Connection) Close() bool {
	if sqlDB, err := c.driver.DB(); err != nil {
		slog.Error("There was an error closing the db: " + err.Error())

		return false
	} else {
		if err = sqlDB.Close(); err != nil {
			slog.Error("There was an error closing the db: " + err.Error())
			return false
		}
	}

	return true
}

func (c *Connection) Ping() {
	var driver *sql.DB

	slog.Info("Database ping started", "separator", "---------")

	if conn, err := c.driver.DB(); err != nil {
		slog.Error("Error retrieving the db driver", "error", err.Error())

		return
	} else {
		driver = conn
		slog.Info("Database driver acquired", "type", fmt.Sprintf("%T", driver))
	}

	if err := driver.Ping(); err != nil {
		slog.Error("Error pinging the db driver", "error", err.Error())
	}

	slog.Info("Database driver is healthy", "stats", driver.Stats())

	slog.Info("Database ping completed", "separator", "---------")
}

func (c *Connection) Sql() *gorm.DB {
	return c.driver
}

func (c *Connection) GetSession() *gorm.Session {
	return &gorm.Session{QueryFields: true}
}

func (c *Connection) Transaction(callback func(db *gorm.DB) error) error {
	return c.driver.Transaction(callback)
}
