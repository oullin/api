package http

import "testing"

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
