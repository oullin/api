package database

import (
	"strings"
	"testing"
)

func TestIsValidTable(t *testing.T) {
	for _, name := range GetSchemaTables() {
		if !isValidTable(name) {
			t.Fatalf("expected %s table to be valid", name)
		}
	}

	if isValidTable("unknown") {
		t.Fatalf("unexpected valid table")
	}
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
		if isValidTable(name) {
			t.Fatalf("%q should be invalid", name)
		}
	}
}
