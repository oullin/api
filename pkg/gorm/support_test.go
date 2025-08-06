package gorm

import (
	"errors"
	stdgorm "gorm.io/gorm"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(stdgorm.ErrRecordNotFound) {
		t.Fatalf("expected true")
	}

	if IsNotFound(nil) {
		t.Fatalf("nil should be false")
	}
}

func TestIsFoundButHasErrors(t *testing.T) {
	if !IsFoundButHasErrors(errors.New("other")) {
		t.Fatalf("expected true")
	}

	if IsFoundButHasErrors(stdgorm.ErrRecordNotFound) {
		t.Fatalf("should be false")
	}
}

func TestHasDbIssues(t *testing.T) {
	if !HasDbIssues(stdgorm.ErrRecordNotFound) {
		t.Fatalf("expected true")
	}

	if !HasDbIssues(errors.New("foo")) {
		t.Fatalf("expected true")
	}

	if HasDbIssues(nil) {
		t.Fatalf("nil should be false")
	}
}
