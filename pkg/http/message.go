package http

import (
	"fmt"
	baseHttp "net/http"
)

func InternalError(msg string) *ApiError {
	return &ApiError{
		Message: fmt.Sprintf("Internal Server Error: %s", msg),
		Status:  baseHttp.StatusInternalServerError,
	}
}
