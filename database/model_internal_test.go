package database

import "testing"

func TestIsValidTable(t *testing.T) {
	if !isValidTable("users") {
		t.Fatalf("expected users table to be valid")
	}
	if isValidTable("unknown") {
		t.Fatalf("unexpected valid table")
	}
}
