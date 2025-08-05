package database

import (
	"strings"
	"testing"
)

func TestIsValidTable(t *testing.T) {
	for _, name := range GetSchemaTables() {
		name := name
		t.Run(name, func(t *testing.T) {
			if !isValidTable(name) {
				t.Errorf("expected table %q to be valid", name)
			}
		})
	}

	t.Run("nonexistent table", func(t *testing.T) {
		if isValidTable("unknown") {
			t.Error(`expected table "unknown" to be invalid`)
		}
	})
}

func TestIsValidTableNonexistentTables(t *testing.T) {
	invalid := []string{
		"",
		"user!@#",
		"user-name",
		"Users",
		"USERS",
		"table123",
		strings.Repeat("x", 256),
		"   ",
	}

	for _, name := range invalid {
		name := name
		t.Run(name, func(t *testing.T) {
			if isValidTable(name) {
				t.Errorf("%q should be invalid", name)
			}
		})
	}
}
