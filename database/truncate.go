package database

import (
	"errors"
	"fmt"
	"github.com/oullin/metal/env"
)

type Truncate struct {
	database *Connection
	env      *env.Environment
}

func MakeTruncate(db *Connection, env *env.Environment) *Truncate {
	return &Truncate{
		database: db,
		env:      env,
	}
}

func (t Truncate) Execute() error {
	if t.env.App.IsProduction() {
		panic("Cannot truncate production environment")
	}

	tables := GetSchemaTables()

	for i := len(tables) - 1; i >= 0; i-- {

		if !isValidTable(tables[i]) {
			return errors.New(fmt.Sprintf("Table '%s' does not exist", tables[i]))
		}

		t.database.Sql().Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", tables[i]))

		fmt.Println(fmt.Sprintf("Table [%s] sucessfully truncated.", tables[i]))
	}

	return nil
}
