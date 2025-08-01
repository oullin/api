package database

import (
	"strings"
	"testing"
)

func TestIsValidTable(t *testing.T) {
	if !isValidTable("users") {
		t.Fatalf("expected users table to be valid")
	}
	if isValidTable("unknown") {
		t.Fatalf("unexpected valid table")
	}
}

func TestIsValidTableEdgeCases(t *testing.T) {
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
