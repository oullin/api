package database

import (
	"errors"
	"fmt"
	"strings"

	"github.com/oullin/metal/env"
)

type Truncate struct {
	database *Connection
	env      *env.Environment
}

func NewTruncate(db *Connection, env *env.Environment) *Truncate {
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

	db := t.database.Sql()

	for i := len(tables) - 1; i >= 0; i-- {
		table := tables[i]

		if !isValidTable(table) {
			errs = append(errs, fmt.Errorf("table '%s' does not exist", table))
			continue
		}

		if !db.Migrator().HasTable(table) {
			fmt.Printf("[db:truncate] skipped table [%s]: table does not exist\n", table)
			continue
		}

		exec := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table))
		if exec.Error != nil {
			if isUndefinedRelationError(exec.Error) {
				fmt.Printf("[db:truncate] skipped table [%s]: %v\n", table, exec.Error)
				continue
			}

			fmt.Printf("[db:truncate] failed to truncate table [%s]: %v\n", table, exec.Error)
			errs = append(errs, fmt.Errorf("truncate table %s: %w", table, exec.Error))
			continue
		}

		fmt.Printf("[db:truncate] truncated table [%s]\n", table)
	}

	if len(errs) > 0 {
		return fmt.Errorf("truncate completed with %d error(s): %w", len(errs), errors.Join(errs...))
	}
	return nil
}

func isUndefinedRelationError(err error) bool {
	return sqlState(err) == "42P01"
}

func sqlState(err error) string {
	if err == nil {
		return ""
	}

	var stateErr interface{ SQLState() string }
	if errors.As(err, &stateErr) {
		return stateErr.SQLState()
	}

	message := err.Error()
	upper := strings.ToUpper(message)
	marker := "(SQLSTATE "
	idx := strings.LastIndex(upper, marker)
	if idx != -1 {
		start := idx + len(marker)
		end := strings.Index(upper[start:], ")")
		if end != -1 {
			return message[start : start+end]
		}
	}

	return ""
}
