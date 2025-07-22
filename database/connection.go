package database

import (
	"database/sql"
	"fmt"
	"github.com/oullin/env"
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

	fmt.Println("\n---------")

	if conn, err := c.driver.DB(); err != nil {
		fmt.Println(fmt.Sprintf("error retrieving the db driver: %v", err.Error()))

		return
	} else {
		driver = conn
		fmt.Println(fmt.Sprintf("db driver adquired: %T", driver))
	}

	if err := driver.Ping(); err != nil {
		slog.Error("error pinging the db driver: " + err.Error())
	}

	fmt.Println(fmt.Sprintf("db driver is healthy: %+v", driver.Stats()))

	fmt.Println("---------")
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
