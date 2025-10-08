package database

import (
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
	var errs []error

	for i := len(tables) - 1; i >= 0; i-- {
		table := tables[i]

		if !isValidTable(table) {
			errs = append(errs, fmt.Errorf("table '%s' does not exist", table))
			continue
		}

		exec := t.database.Sql().Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table))
		if exec.Error != nil {
			fmt.Printf("[db:truncate] failed to truncate table [%s]: %v\n", table, exec.Error)
			errs = append(errs, fmt.Errorf("truncate table %s: %w", table, exec.Error))
			continue
		}

		fmt.Printf("[db:truncate] truncated table [%s]\n", table)
	}

	if len(errs) > 0 {
		return fmt.Errorf("truncate completed with %d error(s): %v", len(errs), errs)
	}
	return nil
}
