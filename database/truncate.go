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
		table := tables[i]

		if !isValidTable(table) {
			return errors.New(fmt.Sprintf("Table '%s' does not exist", table))
		}

		exec := t.database.Sql().Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table))
		if exec.Error != nil {
			fmt.Printf("[db:truncate] failed to truncate table [%s]: %v\n", table, exec.Error)
			continue
		}

		fmt.Printf("[db:truncate] truncated table [%s]\n", table)
	}

	return nil
}
