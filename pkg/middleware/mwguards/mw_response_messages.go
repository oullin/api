package mwguards

import (
	"log/slog"
	baseHttp "net/http"
	"strings"

	"github.com/oullin/pkg/http"
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

func InvalidRequestError(message, logMessage string, data ...map[string]any) *http.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

        slog.Error(logMessage)

	return &http.ApiError{
		Message: message,
		Status:  baseHttp.StatusUnauthorized,
		Data:    normaliseData(data...),
	}
}

func InvalidTokenFormatError(message, logMessage string, data ...map[string]any) *http.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

        slog.Error(logMessage)

	return &http.ApiError{
		Message: message,
		Status:  baseHttp.StatusUnauthorized,
		Data:    normaliseData(data...),
	}
}

func UnauthenticatedError(message, logMessage string, data ...map[string]any) *http.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

        slog.Error(logMessage)

	return &http.ApiError{
		Message: "2- Invalid credentials: " + logMessage,
		Status:  baseHttp.StatusUnauthorized,
		Data:    normaliseData(data...),
	}
}

func RateLimitedError(message, logMessage string, data ...map[string]any) *http.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

        slog.Error(logMessage)

	return &http.ApiError{
		Message: "Too many authentication attempts",
		Status:  baseHttp.StatusTooManyRequests,
		Data:    normaliseData(data...),
	}
}

func NotFound(message, logMessage string, data ...map[string]any) *http.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

        slog.Error(logMessage)

	return &http.ApiError{
		Message: message,
		Status:  baseHttp.StatusNotFound,
		Data:    normaliseData(data...),
	}
}

func TimestampTooOldError(message, logMessage string, data ...map[string]any) *http.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

        slog.Error(logMessage)

	return &http.ApiError{
		Message: "Request timestamp expired",
		Status:  baseHttp.StatusUnauthorized,
		Data:    normaliseData(data...),
	}
}

func TimestampTooNewError(message, logMessage string, data ...map[string]any) *http.ApiError {
	message, logMessage = normaliseMessages(message, logMessage)

        slog.Error(logMessage)

	return &http.ApiError{
		Message: "Request timestamp invalid",
		Status:  baseHttp.StatusUnauthorized,
		Data:    normaliseData(data...),
	}
}
