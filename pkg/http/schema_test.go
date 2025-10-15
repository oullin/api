package http

import (
	"errors"
	"testing"
)

func TestApiErrorError(t *testing.T) {
	e := &ApiError{
		Message: "boom",
		Status:  500,
	}

	if e.Error() != "boom" {
		t.Fatalf("got %s", e.Error())
	}

	var nilErr *ApiError

	if nilErr.Error() != "Internal Server Error" {
		t.Fatalf("nil error wrong")
	}
}

func TestApiErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	e := &ApiError{
		Message: "boom",
		Status:  500,
		Err:     cause,
	}

	if !errors.Is(e, cause) {
		t.Fatalf("expected errors.Is to match the wrapped cause")
	}

	if got := e.Unwrap(); got != cause {
		t.Fatalf("expected unwrap to return the cause")
	}

	var nilErr *ApiError
	if nilErr.Unwrap() != nil {
		t.Fatalf("expected nil unwrap to be nil")
	}
}
