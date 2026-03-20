package endpoint_test

import (
	"errors"
	"testing"

	"github.com/oullin/internal/shared/endpoint"
)

func TestApiErrorError(t *testing.T) {
	e := &endpoint.ApiError{
		Message: "boom",
		Status:  500,
		Err:     errors.New("boom"),
	}

	if e.Error() != "boom" {
		t.Fatalf("expected error message 'boom', got %q", e.Error())
	}

	var nilErr *endpoint.ApiError

	if nilErr.Error() != "Internal Server Error" {
		t.Fatalf("expected nil error to return 'Internal Server Error', got %q", nilErr.Error())
	}
}

func TestApiErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	e := &endpoint.ApiError{
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

	var nilErr *endpoint.ApiError
	if nilErr.Unwrap() != nil {
		t.Fatalf("expected nil error unwrap to return nil")
	}
}
