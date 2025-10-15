package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	baseHttp "net/http"
	"strings"
)

type Response struct {
	etag         string
	cacheControl string
	writer       baseHttp.ResponseWriter
	request      *baseHttp.Request
	headers      func(w baseHttp.ResponseWriter)
}

func MakeResponseFrom(salt string, writer baseHttp.ResponseWriter, request *baseHttp.Request) *Response {
	etag := fmt.Sprintf(
		`"%s"`,
		strings.TrimSpace(salt),
	)

	cacheControl := "public, max-age=3600"

	return &Response{
		writer:       writer,
		request:      request,
		etag:         strings.TrimSpace(etag),
		cacheControl: cacheControl,
		headers: func(w baseHttp.ResponseWriter) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", cacheControl)
			w.Header().Set("ETag", etag)
		},
	}
}

func MakeNoCacheResponse(writer baseHttp.ResponseWriter, request *baseHttp.Request) *Response {
	cacheControl := "no-store"

	return &Response{
		writer:       writer,
		request:      request,
		cacheControl: cacheControl,
		headers: func(w baseHttp.ResponseWriter) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", cacheControl)
		},
	}
}

func (r *Response) WithHeaders(callback func(w baseHttp.ResponseWriter)) {
	callback(r.writer)
}

func (r *Response) RespondOk(payload any) error {
	w := r.writer
	headers := r.headers

	headers(w)
	w.WriteHeader(baseHttp.StatusOK)

	return json.NewEncoder(r.writer).Encode(payload)
}

func (r *Response) HasCache() bool {
	if r.etag == "" {
		return false
	}

	request := r.request

	match := strings.TrimSpace(
		request.Header.Get("If-None-Match"),
	)

	return match == r.etag
}

func (r *Response) RespondWithNotModified() {
	r.writer.WriteHeader(baseHttp.StatusNotModified)
}

func InternalError(msg string) *ApiError {
	message := fmt.Sprintf("Internal server error: %s", msg)

	return &ApiError{
		Message: message,
		Status:  baseHttp.StatusInternalServerError,
		Err:     errors.New(message),
	}
}

func LogInternalError(msg string, err error) *ApiError {
	slog.Error(err.Error(), "error", err)

	return &ApiError{
		Message: fmt.Sprintf("Internal server error: %s", msg),
		Status:  baseHttp.StatusInternalServerError,
		Err:     err,
	}
}

func BadRequestError(msg string) *ApiError {
	message := fmt.Sprintf("Bad request error: %s", msg)

	return &ApiError{
		Message: message,
		Status:  baseHttp.StatusBadRequest,
		Err:     errors.New(message),
	}
}

func LogBadRequestError(msg string, err error) *ApiError {
	slog.Error(err.Error(), "error", err)

	return &ApiError{
		Message: fmt.Sprintf("Bad request error: %s", msg),
		Status:  baseHttp.StatusBadRequest,
		Err:     err,
	}
}

func LogUnauthorisedError(msg string, err error) *ApiError {
	slog.Error(err.Error(), "error", err)

	return &ApiError{
		Message: fmt.Sprintf("Unauthorised request: %s", msg),
		Status:  baseHttp.StatusUnauthorized,
		Err:     err,
	}
}

func UnprocessableEntity(msg string, errs map[string]any) *ApiError {
	message := fmt.Sprintf("Unprocessable entity: %s", msg)

	return &ApiError{
		Message: message,
		Status:  baseHttp.StatusUnprocessableEntity,
		Data:    errs,
		Err:     errors.New(message),
	}
}

func NotFound(msg string) *ApiError {
	message := fmt.Sprintf("Not found error: %s", msg)

	return &ApiError{
		Message: message,
		Status:  baseHttp.StatusNotFound,
		Err:     errors.New(message),
	}
}
