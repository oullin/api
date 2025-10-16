package mwguards

import (
	"log/slog"
	baseHttp "net/http"
	"strings"

	"github.com/oullin/pkg/endpoint"
)

func normaliseData(data ...map[string]any) map[string]any {
	if data == nil || len(data) == 0 {
		return map[string]any{}
	}

	result := make(map[string]any, len(data))
	for _, d := range data {
		for k, v := range d {
			result[k] = v
		}
	}

	return result
}

func normaliseMessages(message, logMessage string) (string, string) {
	message = strings.TrimSpace(message)

	if strings.TrimSpace(logMessage) == "" {
		logMessage = message
	}

	return message, logMessage
}

func InvalidRequestError(message, logMessage string, data ...map[string]any) *endpoint.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

	d := normaliseData(data...)
	slog.Error(logMessage, "data", d)

	return &endpoint.ApiError{
		Message: message,
		Status:  baseHttp.StatusUnauthorized,
		Data:    d,
	}
}

func InvalidTokenFormatError(message, logMessage string, data ...map[string]any) *endpoint.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

	d := normaliseData(data...)
	slog.Error(logMessage, "data", d)

	return &endpoint.ApiError{
		Message: message,
		Status:  baseHttp.StatusUnauthorized,
		Data:    d,
	}
}

func UnauthenticatedError(message, logMessage string, data ...map[string]any) *endpoint.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

	d := normaliseData(data...)
	slog.Error(logMessage, "data", d)

	return &endpoint.ApiError{
		Message: "2- Invalid credentials: " + logMessage,
		Status:  baseHttp.StatusUnauthorized,
		Data:    d,
	}
}

func RateLimitedError(message, logMessage string, data ...map[string]any) *endpoint.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

	d := normaliseData(data...)
	slog.Error(logMessage, "data", d)

	return &endpoint.ApiError{
		Message: "Too many authentication attempts",
		Status:  baseHttp.StatusTooManyRequests,
		Data:    d,
	}
}

func NotFound(message, logMessage string, data ...map[string]any) *endpoint.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

	d := normaliseData(data...)
	slog.Error(logMessage, "data", d)

	return &endpoint.ApiError{
		Message: message,
		Status:  baseHttp.StatusNotFound,
		Data:    d,
	}
}

func TimestampTooOldError(message, logMessage string, data ...map[string]any) *endpoint.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

	d := normaliseData(data...)
	slog.Error(logMessage, "data", d)

	return &endpoint.ApiError{
		Message: "Request timestamp expired",
		Status:  baseHttp.StatusUnauthorized,
		Data:    d,
	}
}

func TimestampTooNewError(message, logMessage string, data ...map[string]any) *endpoint.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

	d := normaliseData(data...)
	slog.Error(logMessage, "data", d)

	return &endpoint.ApiError{
		Message: "Request timestamp invalid",
		Status:  baseHttp.StatusUnauthorized,
		Data:    d,
	}
}
