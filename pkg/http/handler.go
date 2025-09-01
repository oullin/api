package http

import (
	"encoding/json"
	"log/slog"
	baseHttp "net/http"
)

func MakeApiHandler(fn ApiHandler) baseHttp.HandlerFunc {
	return func(w baseHttp.ResponseWriter, r *baseHttp.Request) {
		if err := fn(w, r); err != nil {
			slog.Error("API Error", "message", err.Message, "status", err.Status)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(err.Status)

			resp := ErrorResponse{
				Error:  err.Message,
				Status: err.Status,
				Data:   err.Data,
			}

			if result := json.NewEncoder(w).Encode(resp); result != nil {
				slog.Error("Could not encode error response", "error", result)
			}
		}
	}
}
