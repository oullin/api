package endpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

type Response struct {
	etag         string
	cacheControl string
	writer       http.ResponseWriter
	request      *http.Request
	headers      func(w http.ResponseWriter)
}

func NewResponseWithCache(salt string, maxAgeSeconds int, writer http.ResponseWriter, request *http.Request) *Response {
	// Ensure non-negative cache duration
	if maxAgeSeconds < 0 {
		maxAgeSeconds = 0
	}

	etag := fmt.Sprintf(
		`"%s"`,
		strings.TrimSpace(salt),
	)

	cacheControl := fmt.Sprintf("public, max-age=%d", maxAgeSeconds)

	return &Response{
		writer:       writer,
		request:      request,
		etag:         strings.TrimSpace(etag),
		cacheControl: cacheControl,
		headers: func(w http.ResponseWriter) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", cacheControl)
			w.Header().Set("ETag", etag)
		},
	}
}

func NewResponseFrom(salt string, writer http.ResponseWriter, request *http.Request) *Response {
	return NewResponseWithCache(salt, 3600, writer, request)
}

func NewNoCacheResponse(writer http.ResponseWriter, request *http.Request) *Response {
	cacheControl := "no-store"

	return &Response{
		writer:       writer,
		request:      request,
		cacheControl: cacheControl,
		headers: func(w http.ResponseWriter) {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Cache-Control", cacheControl)
		},
	}
}

func (r *Response) WithHeaders(callback func(w http.ResponseWriter)) {
	callback(r.writer)
}

func (r *Response) RespondOk(payload any) error {
	w := r.writer
	headers := r.headers

	headers(w)
	w.WriteHeader(http.StatusOK)

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
	r.writer.WriteHeader(http.StatusNotModified)
}

func InternalError(msg string) *ApiError {
	message := fmt.Sprintf("Internal server error: %s", msg)

	return &ApiError{
		Message: message,
		Status:  http.StatusInternalServerError,
		Err:     errors.New(message),
	}
}

func LogInternalError(msg string, err error) *ApiError {
	slog.Error(err.Error(), "error", err)

	return &ApiError{
		Message: fmt.Sprintf("Internal server error: %s", msg),
		Status:  http.StatusInternalServerError,
		Err:     err,
	}
}

func BadRequestError(msg string) *ApiError {
	message := fmt.Sprintf("Bad request error: %s", msg)

	return &ApiError{
		Message: message,
		Status:  http.StatusBadRequest,
		Err:     errors.New(message),
	}
}

func LogBadRequestError(msg string, err error) *ApiError {
	slog.Error(err.Error(), "error", err)

	return &ApiError{
		Message: fmt.Sprintf("Bad request error: %s", msg),
		Status:  http.StatusBadRequest,
		Err:     err,
	}
}

func LogUnauthorisedError(msg string, err error) *ApiError {
	slog.Error(err.Error(), "error", err)

	return &ApiError{
		Message: fmt.Sprintf("Unauthorised request: %s", msg),
		Status:  http.StatusUnauthorized,
		Err:     err,
	}
}

func UnprocessableEntity(msg string, errs map[string]any) *ApiError {
	message := fmt.Sprintf("Unprocessable entity: %s", msg)

	return &ApiError{
		Message: message,
		Status:  http.StatusUnprocessableEntity,
		Data:    errs,
		Err:     errors.New(message),
	}
}

func NotFound(msg string) *ApiError {
	message := fmt.Sprintf("Not found error: %s", msg)

	return &ApiError{
		Message: message,
		Status:  http.StatusNotFound,
		Err:     errors.New(message),
	}
}
