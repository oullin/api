package database

import (
	"fmt"
	"log/slog"

	"github.com/oullin/metal/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

func (c *Connection) Ping() error {
	conn, err := c.driver.DB()
	if err != nil {
		return fmt.Errorf("error retrieving the db driver: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("error pinging the db driver: %w", err)
	}

	return nil
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
