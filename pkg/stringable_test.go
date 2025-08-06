package pkg

import (
	"testing"
	"time"
)

func TestStringable_ToLower(t *testing.T) {
	s := MakeStringable(" FooBar ")

	if got := s.ToLower(); got != "foobar" {
		t.Fatalf("expected foobar got %s", got)
	}
}

func TestStringable_ToSnakeCase(t *testing.T) {
	s := MakeStringable("HelloWorldTest")

	if got := s.ToSnakeCase(); got != "hello_world_test" {
		t.Fatalf("expected hello_world_test got %s", got)
	}
}

func TestStringable_ToDatetime(t *testing.T) {
	s := MakeStringable("2024-06-09")
	dt, err := s.ToDatetime()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dt.Year() != 2024 || dt.Month() != time.June || dt.Day() != 9 {
		t.Fatalf("unexpected datetime: %v", dt)
	}
}

func TestStringable_ToDatetimeError(t *testing.T) {
	s := MakeStringable("bad-date")

	if _, err := s.ToDatetime(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestStringable_Dd(t *testing.T) {
	// just ensure it does not panic and prints
	MakeStringable("test").Dd(struct{ X int }{1})
}
