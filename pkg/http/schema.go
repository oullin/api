package http

import baseHttp "net/http"

type ErrorResponse struct {
	Error  string         `json:"error"`
	Status int            `json:"status"`
	Data   map[string]any `json:"data"`
}

type ApiError struct {
	Message string         `json:"message"`
	Status  int            `json:"status"`
	Data    map[string]any `json:"data"`
}

func (e *ApiError) Error() string {
	if e == nil {
		return "Internal Server Error"
	}

	return e.Message
}

type ApiHandler func(baseHttp.ResponseWriter, *baseHttp.Request) *ApiError

type Middleware func(ApiHandler) ApiHandler
